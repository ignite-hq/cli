package modulecreate

import (
	"fmt"
	"path/filepath"

	"github.com/gobuffalo/genny/v2"
	"github.com/gobuffalo/plush/v4"

	"github.com/ignite/cli/v29/ignite/pkg/errors"
	"github.com/ignite/cli/v29/ignite/pkg/gomodulepath"
	"github.com/ignite/cli/v29/ignite/pkg/placeholder"
	"github.com/ignite/cli/v29/ignite/pkg/protoanalysis/protoutil"
	"github.com/ignite/cli/v29/ignite/pkg/xast"
	"github.com/ignite/cli/v29/ignite/pkg/xgenny"
	"github.com/ignite/cli/v29/ignite/pkg/xstrings"
	"github.com/ignite/cli/v29/ignite/templates/field/plushhelpers"
	"github.com/ignite/cli/v29/ignite/templates/module"
	"github.com/ignite/cli/v29/ignite/templates/typed"
)

// NewIBC returns the generator to scaffold the implementation of the IBCModule interface inside a module.
func NewIBC(replacer placeholder.Replacer, opts *CreateOptions) (*genny.Generator, error) {
	var (
		g        = genny.New()
		template = xgenny.NewEmbedWalker(fsIBC, "files/ibc/", opts.AppPath)
	)

	g.RunFn(genesisModify(replacer, opts))
	g.RunFn(genesisTypesModify(replacer, opts))
	g.RunFn(genesisProtoModify(opts))
	g.RunFn(keysModify(replacer, opts))

	if err := g.Box(template); err != nil {
		return g, err
	}

	appModulePath := gomodulepath.ExtractAppPath(opts.ModulePath)

	ctx := plush.NewContext()
	ctx.Set("moduleName", opts.ModuleName)
	ctx.Set("modulePath", opts.ModulePath)
	ctx.Set("appName", opts.AppName)
	ctx.Set("protoVer", opts.ProtoVer)
	ctx.Set("ibcOrdering", opts.IBCOrdering)
	ctx.Set("dependencies", opts.Dependencies)
	ctx.Set("protoPkgName", module.ProtoPackageName(appModulePath, opts.ModuleName, opts.ProtoVer))

	plushhelpers.ExtendPlushContext(ctx)
	g.Transformer(xgenny.Transformer(ctx))
	g.Transformer(genny.Replace("{{protoDir}}", opts.ProtoDir))
	g.Transformer(genny.Replace("{{appName}}", opts.AppName))
	g.Transformer(genny.Replace("{{moduleName}}", opts.ModuleName))
	g.Transformer(genny.Replace("{{protoVer}}", opts.ProtoVer))

	return g, nil
}

func genesisModify(replacer placeholder.Replacer, opts *CreateOptions) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "module/genesis.go")
		f, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		// Genesis init
		templateInit := `%s
k.SetPort(ctx, genState.PortId)
// Only try to bind to port if it is not already bound, since we may already own
// port capability from capability InitGenesis
if k.ShouldBound(ctx, genState.PortId) {
	// module binds to the port on InitChain
	// and claims the returned capability
	err := k.BindPort(ctx, genState.PortId)
	if err != nil {
		panic("could not claim port capability: " + err.Error())
	}
}`
		replacementInit := fmt.Sprintf(templateInit, typed.PlaceholderGenesisModuleInit)
		content := replacer.Replace(f.String(), typed.PlaceholderGenesisModuleInit, replacementInit)

		// Genesis export
		templateExport := `genesis.PortId = k.GetPort(ctx)
%s`
		replacementExport := fmt.Sprintf(templateExport, typed.PlaceholderGenesisModuleExport)
		content = replacer.Replace(content, typed.PlaceholderGenesisModuleExport, replacementExport)

		newFile := genny.NewFileS(path, content)
		return r.File(newFile)
	}
}

func genesisTypesModify(replacer placeholder.Replacer, opts *CreateOptions) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, "x", opts.ModuleName, "types/genesis.go")
		f, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		// Import
		content, err := xast.AppendImports(
			f.String(),
			xast.WithLastNamedImport("host", "github.com/cosmos/ibc-go/v8/modules/core/24-host"),
		)
		if err != nil {
			return err
		}

		// Default genesis
		templateDefault := `PortId: PortID,
%s`
		replacementDefault := fmt.Sprintf(templateDefault, typed.PlaceholderGenesisTypesDefault)
		content = replacer.Replace(content, typed.PlaceholderGenesisTypesDefault, replacementDefault)

		// Validate genesis
		// PlaceholderIBCGenesisTypeValidate
		templateValidate := `if err := host.PortIdentifierValidator(gs.PortId); err != nil {
	return err
}
%s`
		replacementValidate := fmt.Sprintf(templateValidate, typed.PlaceholderGenesisTypesValidate)
		content = replacer.Replace(content, typed.PlaceholderGenesisTypesValidate, replacementValidate)

		newFile := genny.NewFileS(path, content)
		return r.File(newFile)
	}
}

// Modifies genesis.proto to add a new field.
//
// What it depends on:
//   - Existence of a message named 'GenesisState' in genesis.proto.
func genesisProtoModify(opts *CreateOptions) genny.RunFn {
	return func(r *genny.Runner) error {
		path := opts.ProtoFile("genesis.proto")
		f, err := r.Disk.Find(path)
		if err != nil {
			return err
		}
		protoFile, err := protoutil.ParseProtoFile(f)
		if err != nil {
			return err
		}

		// Grab GenesisState and add next (always 2, I gather) available field.
		// TODO: typed.ProtoGenesisStateMessage exists but in subfolder, so we can't use it here, refactor?
		genesisState, err := protoutil.GetMessageByName(protoFile, "GenesisState")
		if err != nil {
			return errors.Errorf("couldn't find message 'GenesisState' in %s: %w", path, err)
		}
		field := protoutil.NewField("port_id", "string", protoutil.NextUniqueID(genesisState))
		protoutil.Append(genesisState, field)

		newFile := genny.NewFileS(path, protoutil.Print(protoFile))
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
	PortKey = collections.NewPrefix("%[1]v-port-")
)`
		replacementPort := fmt.Sprintf(templatePort, opts.ModuleName)
		content = replacer.Replace(content, module.PlaceholderIBCKeysPort, replacementPort)

		newFile := genny.NewFileS(path, content)
		return r.File(newFile)
	}
}

func appIBCModify(replacer placeholder.Replacer, opts *CreateOptions) genny.RunFn {
	return func(r *genny.Runner) error {
		path := filepath.Join(opts.AppPath, module.PathIBCConfigGo)
		f, err := r.Disk.Find(path)
		if err != nil {
			return err
		}

		// Import
		content, err := xast.AppendImports(
			f.String(),
			xast.WithLastNamedImport(
				fmt.Sprintf("%[1]vmodule", opts.ModuleName),
				fmt.Sprintf("%[1]v/x/%[2]v/module", opts.ModulePath, opts.ModuleName),
			),
			xast.WithLastNamedImport(
				fmt.Sprintf("%[1]vmoduletypes", opts.ModuleName),
				fmt.Sprintf("%[1]v/x/%[2]v/types", opts.ModulePath, opts.ModuleName),
			),
		)
		if err != nil {
			return err
		}

		// create IBC module
		templateIBCModule := `%[2]vIBCModule := ibcfee.NewIBCMiddleware(%[2]vmodule.NewIBCModule(app.%[3]vKeeper), app.IBCFeeKeeper)
ibcRouter.AddRoute(%[2]vmoduletypes.ModuleName, %[2]vIBCModule)
%[1]v`
		replacementIBCModule := fmt.Sprintf(
			templateIBCModule,
			module.PlaceholderIBCNewModule,
			opts.ModuleName,
			xstrings.Title(opts.ModuleName),
		)
		content = replacer.Replace(content, module.PlaceholderIBCNewModule, replacementIBCModule)

		newFile := genny.NewFileS(path, content)
		return r.File(newFile)
	}
}
