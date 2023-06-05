package common

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/ignite/cli/ignite/pkg/cache"
	"github.com/ignite/cli/ignite/pkg/cmdrunner"
	"github.com/ignite/cli/ignite/pkg/cmdrunner/step"
	"github.com/ignite/cli/ignite/pkg/cosmosanalysis/module"
	"github.com/ignite/cli/ignite/pkg/cosmosgen"
	"github.com/ignite/cli/ignite/pkg/gomodule"
	"github.com/ignite/cli/ignite/pkg/xfilepath"
	"github.com/ignite/cli/ignite/services/chain"
	"github.com/pkg/errors"
	gomod "golang.org/x/mod/module"
)

const (
	defaultSDKImport     = "github.com/cosmos/cosmos-sdk"
	moduleCacheNamespace = "analyze.setup.module"
)

var protocGlobalInclude = xfilepath.List(
	xfilepath.JoinFromHome(xfilepath.Path("local/include")),
	xfilepath.JoinFromHome(xfilepath.Path(".local/include")),
)

type ModulesInPath struct {
	Path    string
	Modules []module.Module
}
type AllModules struct {
	ModulePaths []ModulesInPath
	Includes    []string
}
type Analyzer struct {
	ctx          context.Context
	appPath      string
	protoDir     string
	sdkImport    string
	appModules   []module.Module
	cacheStorage cache.Storage
	deps         []gomod.Version
	includeDirs  []string
	thirdModules map[string][]module.Module // app dependency-modules pair.
}

func GetModuleList(ctx context.Context, c *chain.Chain) (map[string]string, error) {

	conf, err := c.Config()
	if err != nil {
		return nil, err
	}

	cacheStorage, err := cache.NewStorage(filepath.Join(c.AppPath(), "analyzer_cache.db"))
	if err != nil {
		return nil, err
	}

	if err := cosmosgen.InstallDepTools(ctx, c.AppPath()); err != nil {
		return nil, err
	}

	var errb bytes.Buffer
	if err := cmdrunner.
		New(
			cmdrunner.DefaultStderr(&errb),
			cmdrunner.DefaultWorkdir(c.AppPath()),
		).Run(ctx, step.New(step.Exec("go", "mod", "download"))); err != nil {
		return nil, errors.Wrap(err, errb.String())
	}

	modFile, err := gomodule.ParseAt(c.AppPath())
	g := &Analyzer{
		ctx:          ctx,
		appPath:      c.AppPath(),
		protoDir:     conf.Build.Proto.Path,
		includeDirs:  conf.Build.Proto.ThirdPartyPaths,
		thirdModules: make(map[string][]module.Module),
		cacheStorage: cacheStorage,
	}
	if err != nil {
		return nil, err
	}

	g.sdkImport = defaultSDKImport

	// Check if the Cosmos SDK import path points to a different path
	// and if so change the default one to the new location.
	for _, r := range modFile.Replace {
		if r.Old.Path == defaultSDKImport {
			g.sdkImport = r.New.Path
			break
		}
	}

	// Read the dependencies defined in the `go.mod` file
	g.deps, err = gomodule.ResolveDependencies(modFile)
	if err != nil {
		return nil, err
	}

	// Discover any custom modules defined by the user's app
	g.appModules, err = g.discoverModules(g.appPath, g.protoDir)
	if err != nil {
		return nil, err
	}

	// Go through the Go dependencies of the user's app within go.mod, some of them might be hosting Cosmos SDK modules
	// that could be in use by user's blockchain.
	//
	// Cosmos SDK is a dependency of all blockchains, so it's absolute that we'll be discovering all modules of the
	// SDK as well during this process.
	//
	// Even if a dependency contains some SDK modules, not all of these modules could be used by user's blockchain.
	// this is fine, we can still generate TS clients for those non modules, it is up to user to use (import in typescript)
	// not use generated modules.
	//
	// TODO: we can still implement some sort of smart filtering to detect non used modules by the user's blockchain
	// at some point, it is a nice to have.
	moduleCache := cache.New[ModulesInPath](g.cacheStorage, moduleCacheNamespace)
	for _, dep := range g.deps {
		// Try to get the cached list of modules for the current dependency package
		cacheKey := cache.Key(dep.Path, dep.Version)
		modulesInPath, err := moduleCache.Get(cacheKey)
		if err != nil && !errors.Is(err, cache.ErrorNotFound) {
			return nil, err
		}

		// Discover the modules of the dependency package when they are not cached
		if errors.Is(err, cache.ErrorNotFound) {
			// Get the absolute path to the package's directory
			path, err := gomodule.LocatePath(g.ctx, g.cacheStorage, c.AppPath(), dep)
			if err != nil {
				return nil, err
			}
			/*
				Regardless of any modules in use directly defined in the go module
				add it to the includeDirs for correct proto include resolution
			*/
			g.includeDirs = append(g.includeDirs, path)
			// Discover any modules defined by the package
			modules, err := g.discoverModules(path, "")
			if err != nil {
				return nil, err
			}

			modulesInPath = ModulesInPath{
				Path:    path,
				Modules: modules,
			}

			if err := moduleCache.Put(cacheKey, modulesInPath); err != nil {
				return nil, err
			}
		}

		g.thirdModules[modulesInPath.Path] = append(g.thirdModules[modulesInPath.Path], modulesInPath.Modules...)
	}
	// Perform include resolution AFTER includeDirs has been fully populated
	includePaths, err := g.resolveInclude(c.AppPath())
	if err != nil {
		return nil, err
	}
	var modulelist []ModulesInPath
	modulelist = append(modulelist, ModulesInPath{Path: c.AppPath(), Modules: g.appModules})
	for sourcePath, modules := range g.thirdModules {
		modulelist = append(modulelist, ModulesInPath{Path: sourcePath, Modules: modules})
	}
	var allModules = &AllModules{
		ModulePaths: modulelist,
		Includes:    includePaths,
	}
	ret := make(map[string]string)
	jsonm, _ := json.Marshal(allModules)
	ret["ModuleAnalysis"] = string(jsonm)
	return ret, nil
}

func (g *Analyzer) resolveDependencyInclude() ([]string, error) {
	// Init paths with the global include paths for protoc
	paths, err := protocGlobalInclude()
	if err != nil {
		return nil, err
	}

	// Relative paths to proto directories
	protoDirs := append([]string{g.protoDir}, g.includeDirs...)

	// Create a list of proto import paths for the dependencies.
	// These paths will be available to be imported from the chain app's proto files.
	for rootPath, m := range g.thirdModules {
		// Skip modules without proto files
		if m == nil {
			continue
		}

		// Check each one of the possible proto directory names for the
		// current module and append them only when the directory exists.
		for _, d := range protoDirs {
			p := filepath.Join(rootPath, d)
			f, err := os.Stat(p)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}

				return nil, err
			}

			if f.IsDir() {
				paths = append(paths, p)
			}
		}
	}

	return paths, nil
}

func (g *Analyzer) resolveInclude(path string) (paths []string, err error) {
	// Append chain app's proto paths
	paths = append(paths, filepath.Join(path, g.protoDir))
	for _, p := range g.includeDirs {
		paths = append(paths, filepath.Join(path, p))
	}

	// Append paths for dependencies that have protocol buffer files
	includePaths, err := g.resolveDependencyInclude()
	if err != nil {
		return nil, err
	}

	paths = append(paths, includePaths...)

	return paths, nil
}

func (g *Analyzer) discoverModules(path, protoDir string) ([]module.Module, error) {
	var filteredModules []module.Module

	modules, err := module.Discover(g.ctx, g.appPath, path, protoDir)
	if err != nil {
		return nil, err
	}

	protoPath := filepath.Join(path, g.protoDir)

	for _, m := range modules {
		if !strings.HasPrefix(m.Pkg.Path, protoPath) {
			continue
		}

		filteredModules = append(filteredModules, m)
	}

	return filteredModules, nil
}
