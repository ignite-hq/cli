// Package scaffolder initializes Ignite CLI apps and modifies existing ones
// to add more features in a later time.
package scaffolder

import (
	"context"
	"os"
	"path/filepath"

	"github.com/ignite/cli/ignite/chainconfig"
	sperrors "github.com/ignite/cli/ignite/errors"
	"github.com/ignite/cli/ignite/pkg/cache"
	"github.com/ignite/cli/ignite/pkg/cmdrunner/exec"
	"github.com/ignite/cli/ignite/pkg/cmdrunner/step"
	"github.com/ignite/cli/ignite/pkg/cosmosanalysis"
	"github.com/ignite/cli/ignite/pkg/cosmosgen"
	"github.com/ignite/cli/ignite/pkg/cosmosver"
	"github.com/ignite/cli/ignite/pkg/gocmd"
	"github.com/ignite/cli/ignite/pkg/gomodule"
	"github.com/ignite/cli/ignite/pkg/gomodulepath"
)

// Scaffolder is Ignite CLI app scaffolder.
type Scaffolder struct {
	// Version of the chain
	Version cosmosver.Version

	// path of the app.
	path string

	// modpath represents the go module path of the app.
	modpath gomodulepath.Path
}

// App creates a new scaffolder for an existent app.
func App(path string) (Scaffolder, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return Scaffolder{}, err
	}

	modpath, path, err := gomodulepath.Find(path)
	if err != nil {
		return Scaffolder{}, err
	}
	modfile, err := gomodule.ParseAt(path)
	if err != nil {
		return Scaffolder{}, err
	}
	if err := cosmosanalysis.ValidateGoMod(modfile); err != nil {
		return Scaffolder{}, err
	}

	version, err := cosmosver.Detect(path)
	if err != nil {
		return Scaffolder{}, err
	}

	if !version.IsFamily(cosmosver.Stargate) {
		return Scaffolder{}, sperrors.ErrOnlyStargateSupported
	}

	s := Scaffolder{
		Version: version,
		path:    path,
		modpath: modpath,
	}

	return s, nil
}

func finish(ctx context.Context, cacheStorage cache.Storage, path, gomodPath string) error {
	// FIXME(tb) untagged version of ignite/cli triggers a 404 not found when go
	// mod tidy requests the sumdb, until we understand why, we disable sumdb.
	// related issue:  https://github.com/golang/go/issues/56174
	opt := exec.StepOption(step.Env("GOSUMDB=off"))
	if err := gocmd.ModTidy(ctx, path, opt); err != nil {
		return err
	}
	if err := protoc(ctx, cacheStorage, path, gomodPath); err != nil {
		return err
	}
	return gocmd.Fmt(ctx, path)
}

func protoc(ctx context.Context, cacheStorage cache.Storage, projectPath, gomodPath string) error {
	if err := cosmosgen.InstallDependencies(ctx, projectPath); err != nil {
		return err
	}

	confpath, err := chainconfig.LocateDefault(projectPath)
	if err != nil {
		return err
	}
	conf, err := chainconfig.ParseFile(confpath)
	if err != nil {
		return err
	}

	options := []cosmosgen.Option{
		cosmosgen.WithGoGeneration(gomodPath),
		cosmosgen.IncludeDirs(conf.Build.Proto.ThirdPartyPaths),
	}

	// generate Typescript Client code as well if it is enabled or when the vuex store is being generated
	if conf.Client.Typescript.Path != "" || conf.Client.Vuex.Path != "" {
		tsClientRootPath := filepath.Join(projectPath, chainconfig.TSClientPath(conf))
		if err := os.MkdirAll(tsClientRootPath, 0o766); err != nil {
			return err
		}

		options = append(options,
			cosmosgen.WithTSClientGeneration(
				cosmosgen.TypescriptModulePath(tsClientRootPath),
				tsClientRootPath,
			),
		)
	}

	// generate Vuex code as well if it is enabled.
	if conf.Client.Vuex.Path != "" {
		storeRootPath := filepath.Join(projectPath, conf.Client.Vuex.Path, "generated")

		options = append(options,
			cosmosgen.WithVuexGeneration(
				false,
				cosmosgen.TypescriptModulePath(storeRootPath),
				storeRootPath,
			),
		)
	}
	if conf.Client.OpenAPI.Path != "" {
		options = append(options, cosmosgen.WithOpenAPIGeneration(conf.Client.OpenAPI.Path))
	}

	return cosmosgen.Generate(ctx, cacheStorage, projectPath, conf.Build.Proto.Path, options...)
}
