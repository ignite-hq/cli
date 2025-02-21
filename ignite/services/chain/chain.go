package chain

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"

	chainconfig "github.com/ignite/cli/v29/ignite/config/chain"
	chainconfigv1 "github.com/ignite/cli/v29/ignite/config/chain/v1"
	"github.com/ignite/cli/v29/ignite/pkg/chaincmd"
	chaincmdrunner "github.com/ignite/cli/v29/ignite/pkg/chaincmd/runner"
	"github.com/ignite/cli/v29/ignite/pkg/cliui/colors"
	uilog "github.com/ignite/cli/v29/ignite/pkg/cliui/log"
	"github.com/ignite/cli/v29/ignite/pkg/confile"
	"github.com/ignite/cli/v29/ignite/pkg/cosmosver"
	"github.com/ignite/cli/v29/ignite/pkg/errors"
	"github.com/ignite/cli/v29/ignite/pkg/events"
	"github.com/ignite/cli/v29/ignite/pkg/repoversion"
	"github.com/ignite/cli/v29/ignite/pkg/xexec"
	"github.com/ignite/cli/v29/ignite/pkg/xurl"
	igniteversion "github.com/ignite/cli/v29/ignite/version"
)

const (
	flagPath = "path"
	flagHome = "home"
)

type (
	// Chain provides programmatic access and tools for a Cosmos SDK blockchain.
	Chain struct {
		// app holds info about blockchain app.
		app App

		options chainOptions

		Version cosmosver.Version

		sourceVersion  version
		serveCancel    context.CancelFunc
		serveRefresher chan struct{}
		served         bool

		ev          events.Bus
		logOutputer uilog.Outputer
	}

	// chainOptions holds user given options that overwrites chain's defaults.
	chainOptions struct {
		// chainID is the chain's id.
		chainID string

		// homePath of the chain's config dir.
		homePath string

		// keyring backend used by commands if not specified in configuration
		keyringBackend chaincmd.KeyringBackend

		// checkDependencies checks that cached Go dependencies of the chain have not
		// been modified since they were downloaded.
		checkDependencies bool

		// checkCosmosSDKVersion checks that the app was scaffolded with version of
		// the Cosmos SDK that is supported by Ignite CLI.
		checkCosmosSDKVersion bool

		// printGeneratedPaths prints the output paths of the generated code
		printGeneratedPaths bool

		// path of a custom config file
		ConfigFile string
	}

	version struct {
		tag  string
		hash string
	}

	// Option configures Chain.
	Option func(*Chain)
)

// ID replaces chain's id with given id.
func ID(id string) Option {
	return func(c *Chain) {
		c.options.chainID = id
	}
}

// HomePath replaces chain's configuration home path with given path.
func HomePath(path string) Option {
	return func(c *Chain) {
		c.options.homePath = path
	}
}

// KeyringBackend specifies the keyring backend to use for the chain command.
func KeyringBackend(keyringBackend chaincmd.KeyringBackend) Option {
	return func(c *Chain) {
		c.options.keyringBackend = keyringBackend
	}
}

// ConfigFile specifies a custom config file to use.
func ConfigFile(configFile string) Option {
	return func(c *Chain) {
		c.options.ConfigFile = configFile
	}
}

// WithOutputer sets the CLI outputer for the chain.
func WithOutputer(s uilog.Outputer) Option {
	return func(c *Chain) {
		c.logOutputer = s
	}
}

// CollectEvents collects events from the chain.
func CollectEvents(ev events.Bus) Option {
	return func(c *Chain) {
		c.ev = ev
	}
}

// CheckDependencies checks that cached Go dependencies of the chain have not
// been modified since they were downloaded. Dependencies are checked by
// running `go mod verify`.
func CheckDependencies() Option {
	return func(c *Chain) {
		c.options.checkDependencies = true
	}
}

// CheckCosmosSDKVersion checks that the app was scaffolded with a version of
// the Cosmos SDK that is supported by Ignite CLI.
func CheckCosmosSDKVersion() Option {
	return func(c *Chain) {
		c.options.checkCosmosSDKVersion = true
	}
}

// PrintGeneratedPaths prints the output paths of the generated code.
func PrintGeneratedPaths() Option {
	return func(c *Chain) {
		c.options.printGeneratedPaths = true
	}
}

// New initializes a new Chain with options that its source lives at path.
func New(path string, options ...Option) (*Chain, error) {
	app, err := NewAppAt(path)
	if err != nil {
		return nil, err
	}

	c := &Chain{
		app:            app,
		serveRefresher: make(chan struct{}, 1),
	}

	// Apply the options
	for _, apply := range options {
		apply(c)
	}

	c.sourceVersion, err = c.appVersion()
	if err != nil && !errors.Is(err, git.ErrRepositoryNotExists) {
		return nil, err
	}

	c.Version, err = cosmosver.Detect(c.app.Path)
	if err != nil {
		return nil, err
	}

	if c.options.checkCosmosSDKVersion {
		if err := igniteversion.AssertSupportedCosmosSDKVersion(c.Version); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func NewWithHomeFlags(cmd *cobra.Command, chainOption ...Option) (*Chain, error) {
	var (
		home, _    = cmd.Flags().GetString(flagHome)
		appPath, _ = cmd.Flags().GetString(flagPath)
	)

	absPath, err := filepath.Abs(appPath)
	if err != nil {
		return nil, err
	}

	// Check if custom home is provided
	if home != "" {
		chainOption = append(chainOption, HomePath(home))
	}
	return New(absPath, chainOption...)
}

func (c *Chain) appVersion() (v version, err error) {
	ver, err := repoversion.Determine(c.app.Path)
	if err != nil {
		return version{}, err
	}

	v.hash = ver.Hash
	v.tag = ver.Tag

	return v, nil
}

// RPCPublicAddress points to the public address of Tendermint RPC, this is shared by
// other chains for relayer related actions.
func (c *Chain) RPCPublicAddress() (string, error) {
	rpcAddress := os.Getenv("RPC_ADDRESS")
	if rpcAddress == "" {
		conf, err := c.Config()
		if err != nil {
			return "", err
		}

		validator, err := chainconfig.FirstValidator(conf)
		if err != nil {
			return "", err
		}

		servers, err := validator.GetServers()
		if err != nil {
			return "", err
		}
		rpcAddress = servers.RPC.Address
	}
	return rpcAddress, nil
}

// ConfigPath returns the config path of the chain.
// Empty string means that the chain has no defined config.
func (c *Chain) ConfigPath() string {
	if c.options.ConfigFile != "" {
		return c.options.ConfigFile
	}
	path, err := chainconfig.LocateDefault(c.app.Path)
	if err != nil {
		return ""
	}
	return path
}

// Config returns the config of the chain.
func (c *Chain) Config() (*chainconfig.Config, error) {
	configPath := c.ConfigPath()
	if configPath == "" {
		return chainconfig.DefaultChainConfig(), nil
	}
	return chainconfig.ParseFile(configPath)
}

// ID returns the chain's id.
func (c *Chain) ID() (string, error) {
	// chainID in App has the most priority.
	if c.options.chainID != "" {
		return c.options.chainID, nil
	}

	// otherwise uses defined in config.yml
	chainConfig, err := c.Config()
	if err != nil {
		return "", err
	}
	genid, ok := chainConfig.Genesis["chain_id"]
	if ok {
		return genid.(string), nil
	}

	// use app name by default.
	return c.app.N(), nil
}

// Name returns the chain's name.
func (c *Chain) Name() string {
	return c.app.N()
}

// Binary returns the name of app's default (appd) binary.
func (c *Chain) Binary() (string, error) {
	conf, err := c.Config()
	if err != nil {
		return "", err
	}

	if conf.Build.Binary != "" {
		return conf.Build.Binary, nil
	}

	return c.app.D(), nil
}

// AbsBinaryPath returns the absolute path to the app's binary.
// Returned path includes the binary name.
func (c *Chain) AbsBinaryPath() (string, error) {
	bin, err := c.Binary()
	if err != nil {
		return "", err
	}

	return xexec.ResolveAbsPath(bin)
}

// SetHome sets the chain home directory.
func (c *Chain) SetHome(home string) {
	c.options.homePath = home
}

// Home returns the blockchain node's home dir.
func (c *Chain) Home() (string, error) {
	// check if home is explicitly defined for the app
	home := c.options.homePath
	if home == "" {
		// return default home otherwise
		var err error
		home, err = c.DefaultHome()
		if err != nil {
			return "", err
		}

	}

	// expand environment variables in home
	home = os.ExpandEnv(home)

	return home, nil
}

// AppPath returns the configured App's path.
func (c *Chain) AppPath() string {
	return c.app.Path
}

// DefaultHome returns the blockchain node's default home dir when not specified in the app.
func (c *Chain) DefaultHome() (string, error) {
	// check if home is defined in config
	cfg, err := c.Config()
	if err != nil {
		return "", err
	}
	validator, _ := chainconfig.FirstValidator(cfg)
	if validator.Home != "" {
		expandedHome, err := expandHome(validator.Home)
		if err != nil {
			return "", err
		}
		validator.Home = expandedHome
		return validator.Home, nil
	}

	return c.appHome(), nil
}

// DefaultGentxPath returns default gentx.json path of the app.
func (c *Chain) DefaultGentxPath() (string, error) {
	home, err := c.Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "config/gentx/gentx.json"), nil
}

// GenesisPath returns genesis.json path of the app.
func (c *Chain) GenesisPath() (string, error) {
	home, err := c.Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "config/genesis.json"), nil
}

// GentxsPath returns the directory where gentxs are stored for the app.
func (c *Chain) GentxsPath() (string, error) {
	home, err := c.Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "config/gentx"), nil
}

// AppTOMLPath returns app.toml path of the app.
func (c *Chain) AppTOMLPath() (string, error) {
	home, err := c.Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "config/app.toml"), nil
}

// ConfigTOMLPath returns config.toml path of the app.
func (c *Chain) ConfigTOMLPath() (string, error) {
	home, err := c.Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "config/config.toml"), nil
}

// ClientTOMLPath returns client.toml path of the app.
func (c *Chain) ClientTOMLPath() (string, error) {
	home, err := c.Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "config/client.toml"), nil
}

// KeyringBackend returns the keyring backend chosen for the chain.
func (c *Chain) KeyringBackend() (chaincmd.KeyringBackend, error) {
	// When keyring backend is initialized as a chain
	// option it overrides any configured backends.
	if c.options.keyringBackend != "" {
		return c.options.keyringBackend, nil
	}

	// Try to get keyring backend from the first configured validator
	cfg, err := c.Config()
	if err != nil {
		return "", err
	}

	validator, _ := chainconfig.FirstValidator(cfg)
	if validator.Client != nil {
		if v, ok := validator.Client["keyring-backend"]; ok {
			if backend, ok := v.(string); ok {
				return chaincmd.KeyringBackendFromString(backend)
			}
		}
	}

	// Try to get keyring backend from client.toml config file
	configTOMLPath, err := c.ClientTOMLPath()
	if err != nil {
		return "", err
	}
	cf := confile.New(confile.DefaultTOMLEncodingCreator, configTOMLPath)
	var conf struct {
		KeyringBackend string `toml:"keyring-backend"`
	}
	if err := cf.Load(&conf); err != nil {
		return "", err
	}
	if conf.KeyringBackend != "" {
		return chaincmd.KeyringBackendFromString(conf.KeyringBackend)
	}

	// Use test backend as default when none is configured
	return chaincmd.KeyringBackendTest, nil
}

// Commands returns the runner execute commands on the chain's binary.
func (c *Chain) Commands(ctx context.Context) (chaincmdrunner.Runner, error) {
	id, err := c.ID()
	if err != nil {
		return chaincmdrunner.Runner{}, err
	}

	home, err := c.Home()
	if err != nil {
		return chaincmdrunner.Runner{}, err
	}

	binary, err := c.Binary()
	if err != nil {
		return chaincmdrunner.Runner{}, err
	}

	// Try to make the binary path absolute. This will also
	// find the binary path when the Go bin path is not part
	// of the PATH environment variable.
	binary = xexec.TryResolveAbsPath(binary)

	backend, err := c.KeyringBackend()
	if err != nil {
		return chaincmdrunner.Runner{}, err
	}

	cfg, err := c.Config()
	if err != nil {
		return chaincmdrunner.Runner{}, err
	}

	servers := chainconfigv1.DefaultServers()
	if len(cfg.Validators) > 0 {
		validator, _ := chainconfig.FirstValidator(cfg)
		servers, err = validator.GetServers()
		if err != nil {
			return chaincmdrunner.Runner{}, err
		}
	}

	nodeAddr, err := xurl.TCP(servers.RPC.Address)
	if err != nil {
		return chaincmdrunner.Runner{}, err
	}

	chainCommandOptions := []chaincmd.Option{
		chaincmd.WithChainID(id),
		chaincmd.WithHome(home),
		chaincmd.WithVersion(c.Version),
		chaincmd.WithNodeAddress(nodeAddr),
		chaincmd.WithKeyringBackend(backend),
	}

	cc := chaincmd.New(binary, chainCommandOptions...)

	ccrOptions := []chaincmdrunner.Option{}

	// Enable command output only when CLI verbosity is enabled
	if c.logOutputer != nil && c.logOutputer.Verbosity() == uilog.VerbosityVerbose {
		out := c.logOutputer.NewOutput(c.app.D(), colors.Cyan)
		ccrOptions = append(
			ccrOptions,
			chaincmdrunner.Stdout(out.Stdout()),
			chaincmdrunner.Stderr(out.Stderr()),
		)
	}

	return chaincmdrunner.New(ctx, cc, ccrOptions...)
}

func appBackendSourceWatchPaths(protoDir string) []string {
	return []string{
		"app",
		"cmd",
		"x",
		"third_party",
		protoDir,
	}
}

// expandHome expands a path that may start with "~" and may contain environment variables.
func expandHome(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		// Only replace the first occurrence at the start.
		path = home + strings.TrimPrefix(path, "~")
	}
	return os.ExpandEnv(path), nil
}
