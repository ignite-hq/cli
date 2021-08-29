package modulecreate

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/plush"
	"github.com/gobuffalo/plushgen"
	"github.com/tendermint/starport/starport/pkg/placeholder"
	"github.com/tendermint/starport/starport/pkg/xgenny"
	"github.com/tendermint/starport/starport/pkg/xstrings"
	"github.com/tendermint/starport/starport/templates/module"
)

// NewIBC returns the generator to scaffold the implementation of the IBCModule interface inside a module
func NewIBC(replacer placeholder.Replacer, opts *CreateOptions) (*genny.Generator, error) {
	var (
		g        = genny.New()
		template = xgenny.NewEmbedWalker(fsIBC, "ibc/", opts.AppPath)
	)

	g.RunFn(moduleModify(replacer, opts))
	g.RunFn(genesisModify(replacer, opts))
	g.RunFn(errorsModify(replacer, opts))
	g.RunFn(genesisTypeModify(replacer, opts))
	g.RunFn(genesisProtoModify(replacer, opts))
	g.RunFn(genesisTestsModify(replacer, opts))
	g.RunFn(keysModify(replacer, opts))
	g.RunFn(keeperModify(replacer, opts))

	if err := g.Box(template); err != nil {
		return g, err
	}
	ctx := plush.NewContext()
	ctx.Set("moduleName", opts.ModuleName)
	ctx.Set("modulePath", opts.ModulePath)
	ctx.Set("appName", opts.AppName)
	ctx.Set("ownerName", opts.OwnerName)
	ctx.Set("ibcOrdering", opts.IBCOrdering)
	ctx.Set("title", strings.Title)
	ctx.Set("dependencies", opts.Dependencies)

	// Used for proto package name
	ctx.Set("formatOwnerName", xstrings.FormatUsername)

	g.Transformer(plushgen.Transformer(ctx))
	g.Transformer(genny.Replace("{{moduleName}}", opts.ModuleName))
	return g, nil
}

func moduleModify(replacer placeholder.Replacer, opts *CreateOptions) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "module.go")
		f, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		// Import
		templateImport := `porttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/05-port/types"`
		content := replacer.Replace(f.String(), module.PlaceholderIBCModuleImport, templateImport)

		// Interface to implement
		templateInterface := `_ porttypes.IBCModule   = AppModule{}`
		content = replacer.Replace(content, module.PlaceholderIBCModuleInterface, templateInterface)

		newFile := genny.NewFileS(path, content)
		return r.File(newFile)
	}
}

func genesisModify(replacer placeholder.Replacer, opts *CreateOptions) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "genesis.go")
		f, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		// Genesis init
		templateInit := `k.SetPort(ctx, genState.PortId)
// Only try to bind to port if it is not already bound, since we may already own
// port capability from capability InitGenesis
if !k.IsBound(ctx, genState.PortId) {
	// module binds to the port on InitChain
	// and claims the returned capability
	err := k.BindPort(ctx, genState.PortId)
	if err != nil {
		panic("could not claim port capability: " + err.Error())
	}
}`
		content := replacer.Replace(f.String(), module.PlaceholderIBCGenesisInit, templateInit)

		// Genesis export
		templateExport := `genesis.PortId = k.GetPort(ctx)`
		content = replacer.Replace(content, module.PlaceholderIBCGenesisExport, templateExport)

		newFile := genny.NewFileS(path, content)
		return r.File(newFile)
	}
}

func errorsModify(replacer placeholder.Replacer, opts *CreateOptions) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "types/errors.go")
		f, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		// IBC errors
		template := `ErrInvalidPacketTimeout = sdkerrors.Register(ModuleName, 1500, "invalid packet timeout")
ErrInvalidVersion = sdkerrors.Register(ModuleName, 1501, "invalid version")`
		content := replacer.Replace(f.String(), module.PlaceholderIBCErrors, template)

		newFile := genny.NewFileS(path, content)
		return r.File(newFile)
	}
}

func genesisTypeModify(replacer placeholder.Replacer, opts *CreateOptions) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "types/genesis.go")
		f, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		// Import
		templateImport := `host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"`
		content := replacer.Replace(f.String(), module.PlaceholderIBCGenesisTypeImport, templateImport)

		// Default genesis
		templateDefault := `PortId: PortID,`
		content = replacer.Replace(content, module.PlaceholderIBCGenesisTypeDefault, templateDefault)

		// Validate genesis
		// PlaceholderIBCGenesisTypeValidate
		templateValidate := `if err := host.PortIdentifierValidator(gs.PortId); err != nil {
	return err
}`
		content = replacer.Replace(content, module.PlaceholderIBCGenesisTypeValidate, templateValidate)

		newFile := genny.NewFileS(path, content)
		return r.File(newFile)
	}
}

func genesisProtoModify(replacer placeholder.Replacer, opts *CreateOptions) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "proto", opts.ModuleName, "genesis.proto")
		f, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		// Determine the new field number
		content := f.String()
		fieldNumber := strings.Count(content, module.PlaceholderGenesisProtoStateField) + 1

		template := `string port_id = %[1]v; %[2]v`
		replacement := fmt.Sprintf(template, fieldNumber, module.PlaceholderGenesisProtoStateField)
		content = replacer.Replace(content, module.PlaceholderIBCGenesisProto, replacement)

		newFile := genny.NewFileS(path, content)
		return r.File(newFile)
	}
}

func genesisTestsModify(replacer placeholder.Replacer, opts *CreateOptions) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "types/genesis_test.go")
		f, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		replacement := fmt.Sprintf("PortId: types.PortID,\n%s", module.PlaceholderTypesGenesisValidField)
		content := replacer.Replace(f.String(), module.PlaceholderTypesGenesisValidField, replacement)

		newFile := genny.NewFileS(path, content)
		return r.File(newFile)
	}
}

func keysModify(replacer placeholder.Replacer, opts *CreateOptions) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "types/keys.go")
		f, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		// Append version and the port ID in keys
		templateName := `// Version defines the current version the IBC module supports
Version = "%[1]v-1"

// PortID is the default port id that module binds to
PortID = "%[1]v"`
		replacementName := fmt.Sprintf(templateName, opts.ModuleName)
		content := replacer.Replace(f.String(), module.PlaceholderIBCKeysName, replacementName)

		// PlaceholderIBCKeysPort
		templatePort := `var (
	// PortKey defines the key to store the port ID in store
	PortKey = KeyPrefix("%[1]v-port-")
)`
		replacementPort := fmt.Sprintf(templatePort, opts.ModuleName)
		content = replacer.Replace(content, module.PlaceholderIBCKeysPort, replacementPort)

		newFile := genny.NewFileS(path, content)
		return r.File(newFile)
	}
}

func keeperModify(replacer placeholder.Replacer, opts *CreateOptions) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "keeper/keeper.go")
		f, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		// Keeper new attributes
		templateAttribute := `channelKeeper types.ChannelKeeper
portKeeper    types.PortKeeper
scopedKeeper  types.ScopedKeeper`
		content := replacer.Replace(f.String(), module.PlaceholderIBCKeeperAttribute, templateAttribute)

		// New parameter for the constructor
		templateParameter := `channelKeeper types.ChannelKeeper,
portKeeper types.PortKeeper,
scopedKeeper types.ScopedKeeper,`
		content = replacer.Replace(content, module.PlaceholderIBCKeeperParameter, templateParameter)

		// New return values for the constructor
		templateReturn := `channelKeeper: channelKeeper,
portKeeper:    portKeeper,
scopedKeeper:  scopedKeeper,`
		content = replacer.Replace(content, module.PlaceholderIBCKeeperReturn, templateReturn)

		newFile := genny.NewFileS(path, content)
		return r.File(newFile)
	}
}

func appIBCModify(replacer placeholder.Replacer, opts *CreateOptions) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, module.PathAppGo)
		f, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		// Add route to IBC router
		templateRouter := `%[1]v
ibcRouter.AddRoute(%[2]vmoduletypes.ModuleName, %[2]vModule)`
		replacementRouter := fmt.Sprintf(
			templateRouter,
			module.PlaceholderIBCAppRouter,
			opts.ModuleName,
		)
		content := replacer.Replace(f.String(), module.PlaceholderIBCAppRouter, replacementRouter)

		// Scoped keeper declaration for the module
		templateScopedKeeperDeclaration := `Scoped%[1]vKeeper capabilitykeeper.ScopedKeeper`
		replacementScopedKeeperDeclaration := fmt.Sprintf(templateScopedKeeperDeclaration, strings.Title(opts.ModuleName))
		content = replacer.Replace(content, module.PlaceholderIBCAppScopedKeeperDeclaration, replacementScopedKeeperDeclaration)

		// Scoped keeper definition
		templateScopedKeeperDefinition := `scoped%[1]vKeeper := app.CapabilityKeeper.ScopeToModule(%[2]vmoduletypes.ModuleName)
app.Scoped%[1]vKeeper = scoped%[1]vKeeper`
		replacementScopedKeeperDefinition := fmt.Sprintf(
			templateScopedKeeperDefinition,
			strings.Title(opts.ModuleName),
			opts.ModuleName,
		)
		content = replacer.Replace(content, module.PlaceholderIBCAppScopedKeeperDefinition, replacementScopedKeeperDefinition)

		// New argument passed to the module keeper
		templateKeeperArgument := `app.IBCKeeper.ChannelKeeper,
&app.IBCKeeper.PortKeeper,
scoped%[1]vKeeper,`
		replacementKeeperArgument := fmt.Sprintf(
			templateKeeperArgument,
			strings.Title(opts.ModuleName),
		)
		content = replacer.Replace(content, module.PlaceholderIBCAppKeeperArgument, replacementKeeperArgument)

		newFile := genny.NewFileS(path, content)
		return r.File(newFile)
	}
}
