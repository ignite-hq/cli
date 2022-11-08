package plugin

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/rpc"
	"strconv"

	"github.com/hashicorp/go-plugin"
	"github.com/spf13/pflag"
)

func init() {
	gob.Register(Command{})
	gob.Register(Hook{})
}

// An ignite plugin must implements the Plugin interface.
//
//go:generate mockery --srcpkg . --name Interface --structname PluginInterface --filename interface.go --with-expecter
type Interface interface {
	// Commands returns one or multiple commands that will be added to the list
	// of ignite commands. It's invoked each time ignite is executed, in
	// order to display the list of available commands.
	// Each commands are independent, for nested commands, use the field
	// Command.Commands.
	Commands() ([]Command, error)
	// Execute will be invoked by ignite when a plugin commands is executed.
	// cmd is the executed command (one of the those returned by Commands method)
	// args is the command line arguments passed behing the command.
	Execute(cmd Command, args []string) error

	// Hooks defines custom hooks registered with a given plugin
	Hooks() ([]Hook, error)
	// ExecuteHookPre is invoked by Ignite when a command specified by the hook
	// path is invoked is global for all hooks registered to a plugin context on
	// the hook being invoked is given by the `hook` parameter.
	ExecuteHookPre(hook Hook, args []string) error
	// ExecuteHookPost is invoked by Ignite when a command specified by the hook
	// path is invoked is global for all hooks registered to a plugin context on
	// the hook being invoked is given by the `hook` parameter.
	ExecuteHookPost(hook Hook, args []string) error
	// ExecuteHookCleanUp is invoked right before the command is done executing
	// will be called regardless of execution status of the command and hooks.
	ExecuteHookCleanUp(hook Hook, args []string) error
}

// Command represents a plugin command.
type Command struct {
	// Same as cobra.Command.Use
	Use string
	// Same as cobra.Command.Short
	Short string
	// Same as cobra.Command.Long
	Long string
	// PlaceCommandUnder indicates where the command should be placed.
	// For instance `ignite scaffold` will place the command at the
	// `scaffold` command.
	// An empty value is interpreted as `ignite` (==root).
	PlaceCommandUnder string
	// List of sub commands
	Commands []Command

	// Optionnal parameters populated by config at runtime via
	// chainconfig.Plugin.With field.
	With map[string]string

	flags *pflag.FlagSet
}

func (c *Command) Flags() *pflag.FlagSet {
	if c.flags == nil {
		c.flags = pflag.NewFlagSet(c.Use, pflag.ContinueOnError)
	}
	return c.flags
}

func (c *Command) SetFlags(fs *pflag.FlagSet) {
	c.flags = fs
}

// gobCommandFlags is used to gob encode/decode Command.
// Command can't be encoded because :
// - flags is unexported (because we want to expose it via the Flags() method,
// like a regular cobra.Command)
// - flags type is *pflag.FlagSet which is also full of unexported fields.
type gobCommandFlags struct {
	Command gobCommand
	Flags   []flag
}

type gobCommand Command

type flag struct {
	Name      string // name as it appears on command line
	Shorthand string // one-letter abbreviated flag
	Usage     string // help message
	DefValue  string // default value (as text); for usage message
	Value     string
	Type      flagType
}

type flagType string

const (
	flagTypeString flagType = "string"
	flagTypeInt    flagType = "int"
	flagTypeInt64  flagType = "int64"
	flagTypeBool   flagType = "bool"
)

// GobEncode implements gob.Encoder.
// It actually encodes a gobCommand struct built from c.
func (c Command) GobEncode() ([]byte, error) {
	var ff []flag
	if c.flags != nil {
		c.flags.VisitAll(func(pf *pflag.Flag) {
			ff = append(ff, flag{
				Name:      pf.Name,
				Shorthand: pf.Shorthand,
				Usage:     pf.Usage,
				DefValue:  pf.DefValue,
				Value:     pf.Value.String(),
				Type:      flagType(pf.Value.Type()),
			})
		})
	}
	var b bytes.Buffer
	err := gob.NewEncoder(&b).Encode(gobCommandFlags{
		Command: gobCommand(c),
		Flags:   ff,
	})
	return b.Bytes(), err
}

// GobDecode implements gob.Decoder.
// It actually decodes a gobCommand struct and fills c with it.
func (c *Command) GobDecode(bz []byte) error {
	var gb gobCommandFlags
	err := gob.NewDecoder(bytes.NewReader(bz)).Decode(&gb)
	if err != nil {
		return err
	}
	*c = Command(gb.Command)
	for _, f := range gb.Flags {
		switch f.Type {
		case flagTypeBool:
			defVal, _ := strconv.ParseBool(f.DefValue)
			c.Flags().BoolP(f.Name, f.Shorthand, defVal, f.Usage)
			c.Flags().Set(f.Name, f.Value)
		case flagTypeInt:
			defVal, _ := strconv.Atoi(f.DefValue)
			c.Flags().IntP(f.Name, f.Shorthand, defVal, f.Usage)
			c.Flags().Set(f.Name, f.Value)
		case flagTypeInt64:
			defVal, _ := strconv.ParseInt(f.DefValue, 10, 64)
			c.Flags().Int64P(f.Name, f.Shorthand, defVal, f.Usage)
			c.Flags().Set(f.Name, f.Value)
		case flagTypeString:
			c.Flags().StringP(f.Name, f.Shorthand, f.DefValue, f.Usage)
			c.Flags().Set(f.Name, f.Value)
		default:
			panic(fmt.Sprintf("flagset unmarshal: unhandled flag type %#v", f))
		}
	}
	return nil
}

// Hook represents a user defined action within a plugin
type Hook struct {
	// Name identifies the hook for the client to invoke the correct hook
	// must be unique
	Name string
	// PlaceHookOn indicates the command to register the hooks for
	PlaceHookOn string

	// Optionnal parameters populated by config at runtime via
	// chainconfig.Plugin.With field.
	With map[string]string
}

// handshakeConfigs are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

func HandshakeConfig() plugin.HandshakeConfig {
	return handshakeConfig
}

// Here is an implementation that talks over RPC
type InterfaceRPC struct{ client *rpc.Client }

// Commands implements Interface.Commands
func (g *InterfaceRPC) Commands() ([]Command, error) {
	var resp []Command
	err := g.client.Call("Plugin.Commands", new(interface{}), &resp)
	if err != nil {
		return nil, fmt.Errorf("error while calling plugin Commands: %w", err)
	}
	return resp, nil
}

// Hooks implements Interface.Hooks
func (g *InterfaceRPC) Hooks() ([]Hook, error) {
	var resp []Hook
	err := g.client.Call("Plugin.Hooks", new(interface{}), &resp)
	if err != nil {
		return nil, fmt.Errorf("error while calling plugin Hooks: %w", err)
	}
	return resp, nil
}

// Execute implements Interface.Commands
func (g *InterfaceRPC) Execute(c Command, args []string) error {
	var resp interface{}
	return g.client.Call("Plugin.Execute", map[string]interface{}{
		"command": c,
		"args":    args,
	}, &resp)
}

func (g *InterfaceRPC) ExecuteHookPre(hook Hook, args []string) error {
	var resp interface{}
	return g.client.Call("Plugin.ExecuteHookPre", map[string]interface{}{
		"hook": hook,
		"args": args,
	}, &resp)
}

func (g *InterfaceRPC) ExecuteHookPost(hook Hook, args []string) error {
	var resp interface{}
	return g.client.Call("Plugin.ExecuteHookPost", map[string]interface{}{
		"hook": hook,
		"args": args,
	}, &resp)
}

func (g *InterfaceRPC) ExecuteHookCleanUp(hook Hook, args []string) error {
	var resp interface{}
	return g.client.Call("Plugin.ExecuteHookCleanUp", map[string]interface{}{
		"hook": hook,
		"args": args,
	}, &resp)
}

// Here is the RPC server that InterfaceRPC talks to, conforming to
// the requirements of net/rpc
type InterfaceRPCServer struct {
	// This is the real implementation
	Impl Interface
}

func (s *InterfaceRPCServer) Commands(args interface{}, resp *[]Command) error {
	var err error
	*resp, err = s.Impl.Commands()
	return err
}

func (s *InterfaceRPCServer) Hooks(args interface{}, resp *[]Hook) error {
	var err error
	*resp, err = s.Impl.Hooks()
	return err
}

func (s *InterfaceRPCServer) Execute(args map[string]interface{}, resp *interface{}) error {
	return s.Impl.Execute(args["command"].(Command), args["args"].([]string))
}

func (s *InterfaceRPCServer) ExecuteHookPre(args map[string]interface{}, resp *interface{}) error {
	return s.Impl.ExecuteHookPre(args["hook"].(Hook), args["args"].([]string))
}

func (s *InterfaceRPCServer) ExecuteHookPost(args map[string]interface{}, resp *interface{}) error {
	return s.Impl.ExecuteHookPost(args["hook"].(Hook), args["args"].([]string))
}

func (s *InterfaceRPCServer) ExecuteHookCleanUp(args map[string]interface{}, resp *interface{}) error {
	return s.Impl.ExecuteHookCleanUp(args["hook"].(Hook), args["args"].([]string))
}

// This is the implementation of plugin.Interface so we can serve/consume this
//
// This has two methods: Server must return an RPC server for this plugin
// type. We construct a InterfaceRPCServer for this.
//
// Client must return an implementation of our interface that communicates
// over an RPC client. We return InterfaceRPC for this.
//
// Ignore MuxBroker. That is used to create more multiplexed streams on our
// plugin connection and is a more advanced use case.
type InterfacePlugin struct {
	// Impl Injection
	Impl Interface
}

func (p *InterfacePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &InterfaceRPCServer{Impl: p.Impl}, nil
}

func (InterfacePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &InterfaceRPC{client: c}, nil
}
