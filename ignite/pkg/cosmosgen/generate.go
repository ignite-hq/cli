package cosmosgen

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"gopkg.in/yaml.v3"

	"github.com/ignite/cli/v29/ignite/config/chain/defaults"
	"github.com/ignite/cli/v29/ignite/pkg/cache"
	"github.com/ignite/cli/v29/ignite/pkg/cliui/colors"
	"github.com/ignite/cli/v29/ignite/pkg/cliui/icons"
	"github.com/ignite/cli/v29/ignite/pkg/cmdrunner"
	"github.com/ignite/cli/v29/ignite/pkg/cmdrunner/step"
	"github.com/ignite/cli/v29/ignite/pkg/cosmosanalysis/module"
	"github.com/ignite/cli/v29/ignite/pkg/cosmosbuf"
	"github.com/ignite/cli/v29/ignite/pkg/cosmosver"
	"github.com/ignite/cli/v29/ignite/pkg/errors"
	"github.com/ignite/cli/v29/ignite/pkg/events"
	"github.com/ignite/cli/v29/ignite/pkg/gomodule"
	"github.com/ignite/cli/v29/ignite/pkg/xfilepath"
	"github.com/ignite/cli/v29/ignite/pkg/xos"
)

const (
	moduleCacheNamespace       = "generate.setup.module"
	includeProtoCacheNamespace = "generator.includes.proto"
	workFilename               = "buf.work.yaml"
)

var (
	ErrBufConfig     = errors.New("invalid Buf config")
	ErrMissingSDKDep = errors.New("cosmos-sdk dependency not found")

	protocGlobalInclude = xfilepath.List(
		xfilepath.JoinFromHome(xfilepath.Path("local/include")),
		xfilepath.JoinFromHome(xfilepath.Path(".local/include")),
	)
)

// protoIncludes contains proto include paths for a package.
type protoIncludes struct {
	// Paths is a list of proto include paths.
	Paths []string

	// BufPath is the path to the Buf config file when it exists.
	BufPath string

	// ProtoPath contains the path to the package's proto directory.
	ProtoPath string
}

// protoAnalysis contains proto module analysis data for a Go package dependency.
type protoAnalysis struct {
	// Path is the full path to the Go dependency
	Path string

	// Modules contains the proto modules analysis data.
	// The list is empty when the Go package has no proto files.
	Modules []module.Module

	// Includes contain proto include paths.
	// These paths should be used when generating code.
	Includes protoIncludes
}

func newBufConfigError(path string, cause error) error {
	return errors.Errorf("%w: %s: %w", ErrBufConfig, path, cause)
}

func (g *generator) setup(ctx context.Context) (err error) {
	// Cosmos SDK hosts proto files of own x/ modules and some third party ones needed by itself and
	// blockchain apps. Generate should be aware of these and make them available to the blockchain
	// app that wants to generate code for its own proto.
	//
	// blockchain apps may use different versions of the SDK. following code first makes sure that
	// app's dependencies are download by 'go mod' and cached under the local filesystem.
	// and then, it determines which version of the SDK is used by the app and what is the absolute path
	// of its source code.
	var errb bytes.Buffer
	if err := cmdrunner.
		New(
			cmdrunner.DefaultStderr(&errb),
			cmdrunner.DefaultWorkdir(g.appPath),
		).Run(ctx, step.New(step.Exec("go", "mod", "download"))); err != nil { // TODO: use gocmd.ModDownload
		return errors.Wrap(err, errb.String())
	}

	modFile, err := gomodule.ParseAt(g.appPath)
	if err != nil {
		return err
	}

	// Read the dependencies defined in the `go.mod` file
	g.deps, err = gomodule.ResolveDependencies(modFile, false)
	if err != nil {
		return err
	}

	// Dependencies are resolved, it is possible that the cosmos sdk has been replaced
	g.sdkImport = cosmosver.CosmosModulePath
	for _, dep := range g.deps {
		if cosmosver.CosmosSDKModulePathPattern.MatchString(dep.Path) {
			g.sdkImport = dep.Path
			break
		}
	}

	// Discover any custom modules defined by the user's app.
	// Use the configured proto directory to locate app's proto files.
	g.appModules, err = module.Discover(
		ctx,
		g.appPath,
		g.appPath,
		module.WithProtoDir(g.protoDir),
		module.WithSDKDir(g.sdkDir),
	)
	if err != nil {
		return err
	}

	g.appIncludes, _, err = g.resolveIncludes(ctx, g.appPath, g.protoDir)
	if err != nil {
		return err
	}

	dep, found := filterCosmosSDKModule(g.deps)
	if !found {
		return ErrMissingSDKDep
	}

	// Find the full path to the Cosmos SDK Go package.
	// The path is required to be able to discover proto packages for the
	// set of "cosmossdk.io" packages that doesn't contain the proto files.
	g.sdkDir, err = gomodule.LocatePath(ctx, g.cacheStorage, g.appPath, dep)
	if err != nil {
		return err
	}

	// Go through the Go dependencies of the user's app within go.mod, some of them
	// might be hosting Cosmos SDK modules that could be in use by user's blockchain.
	//
	// Cosmos SDK is a dependency of all blockchains, so it's absolute that we'll be
	// discovering all modules of the SDK as well during this process.
	//
	// Even if a dependency contains some SDK modules, not all of these modules could
	// be used by user's blockchain. This is fine, we can still generate TS clients
	// for those non modules, it is up to user to use (import in typescript) not use
	// generated modules.
	//
	// TODO: we can still implement some sort of smart filtering to detect non used
	// modules by the user's blockchain at some point, it is a nice to have.
	moduleCache := cache.New[protoAnalysis](g.cacheStorage, moduleCacheNamespace)
	for _, dep := range g.deps {
		// Try to get the cached list of modules for the current dependency package
		cacheKey := cache.Key(dep.Path, dep.Version)
		depInfo, err := moduleCache.Get(cacheKey)
		if err != nil && !errors.Is(err, cache.ErrorNotFound) {
			return err
		}

		// Discover the modules of the dependency package when they are not cached
		if errors.Is(err, cache.ErrorNotFound) {
			// Get the absolute path to the package's directory
			path, err := gomodule.LocatePath(ctx, g.cacheStorage, g.appPath, dep)
			if err != nil {
				return err
			}

			// Discover any modules defined by the package.
			// Use an empty string for proto directory because it will be
			// discovered automatically within the dependency package path.
			modules, err := module.Discover(ctx, g.appPath, path, module.WithSDKDir(g.sdkDir))
			if err != nil {
				return err
			}

			// Dependency/includes resolution per module is done to solve versioning issues
			var (
				includes  protoIncludes
				cacheable = true
			)
			if len(modules) > 0 {
				includes, cacheable, err = g.resolveIncludes(ctx, path, defaults.ProtoDir)
				if err != nil {
					return err
				}
			}

			depInfo = protoAnalysis{
				Path:     path,
				Modules:  modules,
				Includes: includes,
			}

			if cacheable {
				if err = moduleCache.Put(cacheKey, depInfo); err != nil {
					return err
				}
			}
		}

		g.thirdModules[depInfo.Path] = depInfo.Modules
		g.thirdModuleIncludes[depInfo.Path] = depInfo.Includes
	}

	return nil
}

func (g *generator) getProtoIncludeFolders(modPath string) []string {
	return []string{filepath.Join(modPath, g.protoDir)}
}

func (g *generator) findBufPath(modpath string) (string, error) {
	var bufPath string
	err := filepath.WalkDir(modpath, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		base := filepath.Base(path)
		if base == "buf.yaml" || base == "buf.yml" {
			bufPath = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return bufPath, nil
}

func (g *generator) generateBufIncludeFolder(ctx context.Context, modpath string) (string, error) {
	protoPath, err := os.MkdirTemp("", "includeFolder")
	if err != nil {
		return "", err
	}

	g.tmpDirs = append(g.tmpDirs, protoPath)

	err = g.buf.Export(ctx, modpath, protoPath)
	if err != nil {
		return "", err
	}
	return protoPath, nil
}

func (g *generator) resolveIncludes(ctx context.Context, path, protoDir string) (protoIncludes, bool, error) {
	// Init paths with the global include paths for protoc
	paths, err := protocGlobalInclude()
	if err != nil {
		return protoIncludes{}, false, err
	}

	includes := protoIncludes{Paths: paths}

	// The "cosmossdk.io" module packages must use SDK's proto path which is
	// where all proto files for there type of Go package are.
	var protoPath string
	if module.IsCosmosSDKModulePkg(path) {
		protoPath = filepath.Join(g.sdkDir, "proto")
	} else {
		// Check that the app/package proto directory exists
		protoPath = filepath.Join(path, protoDir)
		fi, err := os.Stat(protoPath)
		if err != nil && !os.IsNotExist(err) {
			return protoIncludes{}, false, err
		} else if !fi.IsDir() {
			// Just return the global includes when a proto directory doesn't exist
			return includes, true, nil
		}
	}

	// Add app's proto path to the list of proto paths
	includes.Paths = append(includes.Paths, protoPath)
	includes.ProtoPath = protoPath

	// Check if the Buf v1 config file is present into the proto path.
	// We can remove it after the Cosmos-SDK migrate to the buf v2.
	includes.BufPath, err = g.findBufPath(protoPath)
	if err != nil {
		return includes, false, err
	}
	// If it was not found, try to find it in the new Buf v2 project structure at the root of the project.
	if includes.BufPath == "" {
		includes.BufPath, err = g.findBufPath(path)
		if err != nil {
			return includes, false, err
		}
	}

	if includes.BufPath != "" {
		// When a Buf config exists export all protos needed
		// to build the modules to a temporary include folder.
		bufProtoPath, err := g.generateBufIncludeFolder(ctx, protoPath)
		if err != nil && !errors.Is(err, cosmosbuf.ErrProtoFilesNotFound) {
			return protoIncludes{}, false, err
		}

		// Use exported files only when the path contains ".proto" files
		if bufProtoPath != "" {
			includes.Paths = append(includes.Paths, bufProtoPath)
			return includes, false, nil
		}
	}

	// When there is no Buf config add the configured directories
	// instead to keep the legacy (non Buf) behavior.
	includes.Paths = append(includes.Paths, g.getProtoIncludeFolders(path)...)

	return includes, true, nil
}

func (g generator) updateBufModule(ctx context.Context) error {
	for pkgPath, includes := range g.thirdModuleIncludes {
		// Skip third party dependencies without proto files
		if includes.ProtoPath == "" {
			continue
		}

		// Resolve the Go package and use the module name as the proto vendor directory name
		modFile, err := gomodule.ParseAt(pkgPath)
		if err != nil {
			return err
		}

		pkgName := modFile.Module.Mod.Path

		// When a Buf config with name is available add it to app's dependencies
		// or otherwise export the proto files to a vendor directory.
		if includes.BufPath != "" {
			if err := g.resolveBufDependency(pkgName, includes.BufPath); err != nil {
				return err
			}
		} else {
			if err := g.vendorProtoPackage(pkgName, includes.ProtoPath); err != nil {
				return err
			}
		}
	}
	if err := g.buf.Update(
		ctx,
		filepath.Dir(g.appIncludes.BufPath),
	); err != nil && !errors.Is(err, cosmosbuf.ErrProtoFilesNotFound) {
		return err
	}
	return nil
}

func (g generator) resolveBufDependency(pkgName, bufPath string) error {
	// Open the dependency Buf config to find the BSR package name
	f, err := os.Open(bufPath)
	if err != nil {
		return err
	}
	defer f.Close()

	cfg := struct {
		Name string `yaml:"name"`
	}{}

	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return newBufConfigError(bufPath, err)
	}

	// When dependency package has a Buf config name try to add it to app's
	// dependencies. Name is optional and defines the BSR package name.
	if cfg.Name != "" {
		return g.addBufDependency(cfg.Name)
	}
	// By default just vendor the proto package
	return g.vendorProtoPackage(pkgName, filepath.Dir(bufPath))
}

func (g generator) addBufDependency(depName string) error {
	// Read app's Buf config
	path := g.appIncludes.BufPath
	bz, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Check if the proto dependency is already present in app's Buf config
	cfg := struct {
		Deps []string `yaml:"deps"`
	}{}
	if err := yaml.Unmarshal(bz, &cfg); err != nil {
		return newBufConfigError(path, err)
	}

	if slices.Contains(cfg.Deps, depName) {
		return nil
	}

	// Add the new dependency and update app's Buf config
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	var rawCfg map[string]interface{}
	if err := yaml.Unmarshal(bz, &rawCfg); err != nil {
		return newBufConfigError(path, err)
	}

	rawCfg["deps"] = append(cfg.Deps, depName)

	enc := yaml.NewEncoder(f)
	defer enc.Close()

	if err := enc.Encode(rawCfg); err != nil {
		return err
	}

	g.opts.ev.Send(
		fmt.Sprintf("New Buf dependency added: %s", colors.Name(depName)),
		events.Icon(icons.OK),
	)

	// Update Buf lock so it contains the new dependency
	return nil
}

func (g generator) vendorProtoPackage(pkgName, protoPath string) (err error) {
	// Check that the dependency vendor directory doesn't exist
	vendorRelPath := filepath.Join("proto_vendor", pkgName)
	vendorPath := filepath.Join(g.appPath, vendorRelPath)
	_, err = os.Stat(vendorPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Skip vendoring when the dependency is already vendored
	if !os.IsNotExist(err) {
		return nil
	}

	if err = os.MkdirAll(vendorPath, 0o777); err != nil {
		return err
	}

	// Make sure that the vendor folder is removed on error
	defer func() {
		if err != nil {
			_ = os.RemoveAll(vendorPath)
		}
	}()

	if err = xos.CopyFolder(protoPath, vendorPath); err != nil {
		return err
	}

	path := filepath.Join(g.appPath, workFilename)
	bz, err := os.ReadFile(path)
	if err != nil {
		return errors.Errorf("error reading Buf workspace file: %s: %w", path, err)
	}

	ws := struct {
		Version     string   `yaml:"version"`
		Directories []string `yaml:"directories"`
	}{}
	if err := yaml.Unmarshal(bz, &ws); err != nil {
		return err
	}

	ws.Directories = append(ws.Directories, vendorRelPath)

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := yaml.NewEncoder(f)
	defer enc.Close()
	if err = enc.Encode(ws); err != nil {
		return err
	}

	g.opts.ev.Send(
		fmt.Sprintf("New Buf vendored dependency added: %s", colors.Name(vendorRelPath)),
		events.Icon(icons.OK),
	)

	return nil
}

func filterCosmosSDKModule(versions []gomodule.Version) (gomodule.Version, bool) {
	for _, v := range versions {
		if cosmosver.CosmosSDKModulePathPattern.MatchString(v.Path) {
			return v, true
		}
	}
	return gomodule.Version{}, false
}
