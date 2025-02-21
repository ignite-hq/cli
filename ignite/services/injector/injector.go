package injector

import (
	"context"

	"github.com/gobuffalo/genny/v2"
	"github.com/gobuffalo/plush/v4"

	"github.com/ignite/cli/v29/ignite/pkg/errors"
	"github.com/ignite/cli/v29/ignite/pkg/gocmd"
	"github.com/ignite/cli/v29/ignite/pkg/xgenny"
	"github.com/ignite/cli/v29/ignite/services/chain"
	"github.com/ignite/cli/v29/ignite/templates/field/plushhelpers"
)

// Injector allows to add modules, commands and other changes to the chain, using a convinient API.
// Once the wanted changes are added, the Inject method should be called to apply them.
type Injector interface {
	AddModule(ctx context.Context, m *Module) error
	AddNonDepinjectModule(ctx context.Context, m *Module) error
	AddCommand(ctx context.Context, c *Command) error

	Inject(ctx context.Context) error
}

// TODO: somewhere else, injector for modules to replace placeholders

type injector struct {
	chain     *chain.Chain
	generator *genny.Generator
}

// NewInject creates an Injector.
func NewInjector(c *chain.Chain) Injector {
	g := genny.New()
	ctx := plush.NewContext()
	plushhelpers.ExtendPlushContext(ctx)
	g.Transformer(xgenny.Transformer(ctx))

	return &injector{
		chain:     c,
		generator: g,
	}
}

// Inject runs the injector, adding all the changes to the chain.
func (i *injector) Inject(ctx context.Context) error {
	appPath := i.chain.AppPath()

	runner := xgenny.NewRunner(ctx, appPath)
	if err := runner.Run(i.generator); err != nil {
		return errors.Errorf("failed to execute generator: %w", err)
	}

	if err := gocmd.ModTidy(ctx, appPath); err != nil {
		return err
	}

	if err := gocmd.Fmt(ctx, appPath); err != nil {
		return err
	}

	_ = gocmd.GoImports(ctx, appPath) // goimports installation could fail, so ignore the error

	return nil
}
