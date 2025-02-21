package injector

import (
	"context"
	"fmt"
	"go/ast"
	"io"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/genny/v2"
	"github.com/ignite/cli/v29/ignite/pkg/cmdrunner/exec"
	"github.com/ignite/cli/v29/ignite/pkg/cmdrunner/step"
	"github.com/ignite/cli/v29/ignite/pkg/gocmd"
	"github.com/ignite/cli/v29/ignite/pkg/goenv"
	"github.com/ignite/cli/v29/ignite/pkg/xast"
)

// Modules should follow the recommended modules structure.
// See https://docs.cosmos.network/main/build/building-modules/structure#structure
type Module struct {
	Name        string // bank
	Import      string // cosmossdk.io/bank
	Permissions string // TODO
}

type moduleConfig struct {
	typePkg   string
	keeperPkg string
	modulePkg string
	apiPkg    string

	hasGenesis      bool // TODO
	hasPreBlocker   bool
	hasBeginBlocker bool
	hasEndBlocker   bool
}

// downloadModule downloads the module and returns its path
func (i *injector) downloadModule(ctx context.Context, m *Module) (string, error) {
	// TODO: allow to check local modules that aren't their own go modules
	// inspect if module is in the chain.

	if err := gocmd.ModDownload(ctx, m.Import, false, exec.StepOption(step.Stdout(io.Discard))); err != nil {
		return "", fmt.Errorf("failed to download module: %w", err)
	}

	return filepath.Join(append([]string{goenv.GoModCache()}, strings.Split(m.Import, "/")...)...), nil
}

// AddModule adds a new module to the chain.
func (i *injector) AddModule(ctx context.Context, m *Module) error {
	mc := moduleConfig{
		typePkg:   fmt.Sprintf("%s/types", m.Import),
		keeperPkg: fmt.Sprintf("%s/keeper", m.Import),
		modulePkg: fmt.Sprintf("%s/module", m.Import),
		apiPkg:    fmt.Sprintf("%s/api", m.Import),
	}

	modulePath, err := i.downloadModule(ctx, m)
	if err != nil {
		return err
	}

	modulePkg, _, err := xast.ParseDir(modulePath)
	if err != nil {
		return err
	}

	for _, f := range modulePkg.Files {
		ast.Inspect(f, func(x ast.Node) bool {
			// TODO find ModuleName and set moduleTypePackage
			// TODO check depinject config and find api module package
			// TODO check abci.go or module.go and find abci implementations
			return true
		})
	}

	appConfigFn := func(r *genny.Runner) error {
		path := i.appConfigPath()

		f, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		content, err := xast.AppendImports(
			f.String(),
			xast.WithLastNamedImport(fmt.Sprintf("%stypes", m.Name), mc.typePkg),
		)
		if err != nil {
			return err
		}

		return r.File(genny.NewFileS(path, content))
	}

	appFn := func(r *genny.Runner) error {
		path := i.appPath()

		f, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		content, err := xast.AppendImports(
			f.String(),
			xast.WithLastNamedImport(fmt.Sprintf("%skeeper", m.Name), mc.keeperPkg),
			xast.WithLastNamedImport("_", mc.modulePkg),
		)
		if err != nil {
			return err
		}

		return r.File(genny.NewFileS(path, content))
	}

	i.generator.RunFn(appConfigFn)
	i.generator.RunFn(appFn)

	return nil
}

func (i *injector) AddNonDepinjectModule(ctx context.Context, m *Module) error {
	// append app.go or create custom.go
	// do not break ibc.go

	panic("not implemented")
}
