package ignitecmd

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/ignite/cli/v29/ignite/config"
	chainconfig "github.com/ignite/cli/v29/ignite/config/chain"
	"github.com/ignite/cli/v29/ignite/pkg/cache"
	"github.com/ignite/cli/v29/ignite/pkg/cliui"
	uilog "github.com/ignite/cli/v29/ignite/pkg/cliui/log"
	"github.com/ignite/cli/v29/ignite/pkg/dircache"
	"github.com/ignite/cli/v29/ignite/pkg/errors"
	"github.com/ignite/cli/v29/ignite/pkg/goenv"
	"github.com/ignite/cli/v29/ignite/pkg/gomodulepath"
	"github.com/ignite/cli/v29/ignite/version"
)

type key int

const (
	keyChainConfig     key = iota
	keyChainConfigPath key = iota
)

const (
	flagPath       = "path"
	flagHome       = "home"
	flagYes        = "yes"
	flagClearCache = "clear-cache"
	flagSkipProto  = "skip-proto"
	flagSkipBuild  = "skip-build"

	checkVersionTimeout = time.Millisecond * 600
	cacheFileName       = "ignite_cache.db"

	statusGenerating = "Generating..."
	statusQuerying   = "Querying..."
)

// List of CLI level one commands that should not load Ignite app instances.
var skipAppsLoadCommands = []string{"version", "help", "docs", "completion", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd}

// New creates a new root command for `Ignite CLI` with its sub commands.
// Returns the cobra.Command, a cleanup function and an error. The cleanup
// function must be invoked by the caller to clean eventual Ignite App instances.
func New(ctx context.Context) (*cobra.Command, func(), error) {
	cobra.EnableCommandSorting = false

	c := &cobra.Command{
		Use:   "ignite",
		Short: "Ignite CLI offers everything you need to scaffold, test, build, and launch your blockchain",
		Long: `Ignite CLI is a tool for creating sovereign blockchains built with Cosmos SDK, the world's
most popular modular blockchain framework. Ignite CLI offers everything you need to scaffold,
test, build, and launch your blockchain.

To get started, create a blockchain:

	ignite scaffold chain example
`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(0), // note(@julienrbrt): without this, ignite __complete(noDesc) hidden commands are not working.
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// Check for new versions only when shell completion scripts are not being
			// generated to avoid invalid output to stdout when a new version is available
			if cmd.Use != "completion" || !strings.HasPrefix(cmd.Use, cobra.ShellCompRequestCmd) {
				checkNewVersion(cmd)
			}

			return goenv.ConfigurePath()
		},
	}

	c.AddCommand(
		NewScaffold(),
		NewChain(),
		NewGenerate(),
		NewAccount(),
		NewDocs(),
		NewVersion(),
		NewApp(),
		NewDoctor(),
		NewCompletionCmd(),
		NewTestnet(),
	)
	c.AddCommand(deprecated()...)
	c.SetContext(ctx)

	// Don't load Ignite apps for level one commands that doesn't allow them
	if len(os.Args) >= 2 && slices.Contains(skipAppsLoadCommands, os.Args[1]) {
		return c, func() {}, nil
	}

	// Load plugins if any
	session := cliui.New(cliui.WithStdout(os.Stdout))
	if err := LoadPlugins(ctx, c, session); err != nil {
		return nil, nil, errors.Errorf("error while loading apps: %w", err)
	}
	return c, func() {
		UnloadPlugins()
		session.End()
	}, nil
}

func flagSetVerbose() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.BoolP(flagVerbose, "v", false, "verbose output")
	return fs
}

func getVerbosity(cmd *cobra.Command) uilog.Verbosity {
	if verbose, _ := cmd.Flags().GetBool(flagVerbose); verbose {
		return uilog.VerbosityVerbose
	}

	return uilog.VerbosityDefault
}

func flagSetPath(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(flagPath, "p", ".", "path of the app")
}

func flagGetPath(cmd *cobra.Command) (path string) {
	path, _ = cmd.Flags().GetString(flagPath)
	return
}

func goModulePath(cmd *cobra.Command) (string, error) {
	path := flagGetPath(cmd)
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	_, appPath, err := gomodulepath.Find(path)
	if err != nil {
		return "", err
	}
	return appPath, err
}

func flagSetHome() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(flagHome, "", "directory where the blockchain node is initialized")
	return fs
}

func flagSetConfig() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.StringP(flagConfig, "c", "", "path to Ignite config file (default: ./config.yml)")
	return fs
}

func getConfig(cmd *cobra.Command) (config string) {
	config, _ = cmd.Flags().GetString(flagConfig)
	return
}

func getChainConfig(cmd *cobra.Command) (*chainconfig.Config, string, error) {
	cfg, ok := cmd.Context().Value(keyChainConfig).(*chainconfig.Config)
	if ok {
		configPath := cmd.Context().Value(keyChainConfigPath).(string)
		return cfg, configPath, nil
	}
	configPath := getConfig(cmd)

	path, err := goModulePath(cmd)
	if err != nil {
		return nil, "", err
	}

	if configPath == "" {
		if configPath, err = chainconfig.LocateDefault(path); err != nil {
			return nil, "", err
		}
	}

	cfg, err = chainconfig.ParseFile(configPath)
	if err != nil {
		return nil, "", err
	}
	ctx := context.WithValue(cmd.Context(), keyChainConfig, cfg)
	ctx = context.WithValue(ctx, keyChainConfigPath, configPath)
	cmd.SetContext(ctx)

	return cfg, configPath, err
}

func flagSetYes() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.BoolP(flagYes, "y", false, "answers interactive yes/no questions with yes")
	return fs
}

func getYes(cmd *cobra.Command) (ok bool) {
	ok, _ = cmd.Flags().GetBool(flagYes)
	return
}

func flagSetSkipProto() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Bool(flagSkipProto, false, "skip file generation from proto")
	return fs
}

func flagGetSkipProto(cmd *cobra.Command) bool {
	skip, _ := cmd.Flags().GetBool(flagSkipProto)
	return skip
}

func flagSetSkipBuild() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Bool(flagSkipBuild, false, "skip initial build of the app (uses local binary)")
	return fs
}

func flagGetSkipBuild(cmd *cobra.Command) bool {
	skip, _ := cmd.Flags().GetBool(flagSkipBuild)
	return skip
}

func flagSetClearCache(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool(flagClearCache, false, "clear the build cache (advanced)")
}

func flagGetClearCache(cmd *cobra.Command) bool {
	clearCache, _ := cmd.Flags().GetBool(flagClearCache)
	return clearCache
}

func deprecated() []*cobra.Command {
	return []*cobra.Command{
		{
			Use:        "build",
			Deprecated: "use `ignite chain build` instead.",
		},
		{
			Use:        "serve",
			Deprecated: "use `ignite chain serve` instead.",
		},
		{
			Use:        "faucet",
			Deprecated: "use `ignite chain faucet` instead.",
		},
		{
			Use:        "node",
			Deprecated: "use ignite connect app instead (ignite app install -g github.com/ignite/apps/connect).",
		},
	}
}

func checkNewVersion(cmd *cobra.Command) {
	ctx, cancel := context.WithTimeout(cmd.Context(), checkVersionTimeout)
	defer cancel()

	isAvailable, next, err := version.CheckNext(ctx)
	if err != nil || !isAvailable {
		return
	}

	cmd.Printf("⬆️ Ignite CLI %s is available! To upgrade: https://docs.ignite.com/welcome/install#upgrade (or use snap or homebrew)\n\n", next)
}

func newCache(cmd *cobra.Command) (cache.Storage, error) {
	cacheRootDir, err := config.DirPath()
	if err != nil {
		return cache.Storage{}, err
	}

	storage, err := cache.NewStorage(
		filepath.Join(cacheRootDir, cacheFileName),
		cache.WithVersion(version.Version),
	)
	if err != nil {
		return cache.Storage{}, err
	}

	if flagGetClearCache(cmd) {
		if err := storage.Clear(); err != nil {
			return cache.Storage{}, err
		}
		if err := dircache.ClearCache(); err != nil {
			return cache.Storage{}, err
		}
	}

	return storage, nil
}
