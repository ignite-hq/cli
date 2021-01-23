package chain

import (
	"bytes"
	"context"
	"fmt"
	"go/build"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/tendermint/starport/starport/services"

	"github.com/otiai10/copy"
	"github.com/pkg/errors"
	conf "github.com/tendermint/starport/starport/chainconf"
	secretconf "github.com/tendermint/starport/starport/chainconf/secret"
	"github.com/tendermint/starport/starport/pkg/cmdrunner"
	"github.com/tendermint/starport/starport/pkg/cmdrunner/step"
	"github.com/tendermint/starport/starport/pkg/cosmosfaucet"
	"github.com/tendermint/starport/starport/pkg/fswatcher"
	"github.com/tendermint/starport/starport/pkg/xexec"
	"github.com/tendermint/starport/starport/pkg/xhttp"
	"github.com/tendermint/starport/starport/pkg/xos"
	"github.com/tendermint/starport/starport/pkg/xurl"
	"golang.org/x/sync/errgroup"
)

var (
	// ignoredExts holds a list of ignored files from watching.
	ignoredExts = []string{"pb.go", "pb.gw.go"}

	// chainSavePath is the place where chain exported genesis are saved
	chainSavePath = filepath.Join(services.StarportConfDir, "local-chains")

	// exportedGenesis is the name of the exported genesis file for a chain
	exportedGenesis = "exported_genesis.json"
)

// Serve serves an app.
func (c *Chain) Serve(ctx context.Context) error {
	// initial checks and setup.
	if err := c.setup(ctx); err != nil {
		return err
	}

	// make sure that config.yml exists.
	if _, err := conf.Locate(c.app.Path); err != nil {
		return err
	}

	// initialize the relayer if application supports it so, secret.yml
	// can be generated and watched for changes.
	if err := c.checkIBCRelayerSupport(); err == nil {
		if _, err := c.RelayerInfo(); err != nil {
			return err
		}
	}

	// start serving components.
	g, ctx := errgroup.WithContext(ctx)

	// routine to watch front-end
	g.Go(func() error {
		return c.watchAppFrontend(ctx)
	})

	// development server routine
	g.Go(func() error {
		return c.runDevServer(ctx)
	})

	// blockchain node routine
	g.Go(func() error {
		c.refreshServe()
		for {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			select {
			case <-ctx.Done():
				return ctx.Err()

			case <-c.serveRefresher:
				var (
					serveCtx context.Context
					buildErr *CannotBuildAppError
				)
				serveCtx, c.serveCancel = context.WithCancel(ctx)

				// serve the app.
				err := c.serve(serveCtx)
				switch {
				case err == nil:
				case errors.Is(err, context.Canceled):
					// If the app has been served, we save the genesis state
					if c.served {
						c.served = false

						fmt.Fprintf(c.stdLog(logStarport).out, "\nSaving genesis state...\n")

						// If serve has been stopped, save the genesis state
						if err := c.saveChainState(context.TODO()); err != nil {
							fmt.Fprint(c.stdLog(logStarport).err, err.Error())
							return err
						}

						genesisPath, err := c.exportedGenesisPath()
						if err != nil {
							fmt.Fprint(c.stdLog(logStarport).err, err.Error())
							return err
						}
						fmt.Fprintf(c.stdLog(logStarport).out, "\nGenesis state saved in %s\n", genesisPath)
					}
				case errors.As(err, &buildErr):
					fmt.Fprintf(c.stdLog(logStarport).err, "%s\n", errorColor(err.Error()))

					var validationErr *conf.ValidationError
					if errors.As(err, &validationErr) {
						fmt.Fprintln(c.stdLog(logStarport).out, "see: https://github.com/tendermint/starport#configure")
					}

					fmt.Fprintf(c.stdLog(logStarport).out, "%s\n", infoColor("waiting for a fix before retrying..."))
				default:
					return err
				}
			}
		}
	})

	// routine to watch back-end
	g.Go(func() error {
		return c.watchAppBackend(ctx)
	})

	return g.Wait()
}

func (c *Chain) setup(ctx context.Context) error {
	fmt.Fprintf(c.stdLog(logStarport).out, "Cosmos SDK's version is: %s\n\n", infoColor(c.Version))

	if err := c.checkSystem(); err != nil {
		return err
	}
	if err := c.plugin.Setup(ctx); err != nil {
		return err
	}
	return nil
}

// checkSystem checks if developer's work environment comply must to have
// dependencies and pre-conditions.
func (c *Chain) checkSystem() error {
	// check if Go has installed.
	if !xexec.IsCommandAvailable("go") {
		return errors.New("Please, check that Go language is installed correctly in $PATH. See https://golang.org/doc/install")
	}
	// check if Go's bin added to System's path.
	gobinpath := path.Join(build.Default.GOPATH, "bin")
	if err := xos.IsInPath(gobinpath); err != nil {
		return errors.New("$(go env GOPATH)/bin must be added to your $PATH. See https://golang.org/doc/gopath_code.html#GOPATH")
	}
	return nil
}

func (c *Chain) refreshServe() {
	if c.serveCancel != nil {
		c.serveCancel()
	}
	c.serveRefresher <- struct{}{}
}

func (c *Chain) watchAppBackend(ctx context.Context) error {
	return fswatcher.Watch(
		ctx,
		appBackendWatchPaths,
		fswatcher.Workdir(c.app.Path),
		fswatcher.OnChange(c.refreshServe),
		fswatcher.IgnoreHidden(),
		fswatcher.IgnoreExt(ignoredExts...),
	)
}

func (c *Chain) cmdOptions() []cmdrunner.Option {
	return []cmdrunner.Option{
		cmdrunner.DefaultWorkdir(c.app.Path),
	}
}

func (c *Chain) serve(ctx context.Context) error {
	conf, err := c.Config()
	if err != nil {
		return &CannotBuildAppError{err}
	}
	sconf, err := secretconf.Open(c.app.Path)
	if err != nil {
		return err
	}

	if err := c.buildProto(ctx); err != nil {
		return err
	}

	buildSteps, err := c.buildSteps()
	if err != nil {
		return err
	}
	if err := cmdrunner.
		New(c.cmdOptions()...).
		Run(ctx, buildSteps...); err != nil {
		return err
	}

	if err := c.Init(ctx); err != nil {
		return err
	}

	for _, account := range conf.Accounts {
		acc, err := c.Commands().AddAccount(ctx, account.Name, "")
		if err != nil {
			return err
		}

		coins := strings.Join(account.Coins, ",")
		if err := c.Commands().AddGenesisAccount(ctx, acc.Address, coins); err != nil {
			return err
		}

		fmt.Fprintf(c.stdLog(logStarport).out, "🙂 Created an account. Password (mnemonic): %[1]v\n", acc.Mnemonic)
	}

	for _, account := range sconf.Accounts {
		acc, err := c.Commands().AddAccount(ctx, account.Name, account.Mnemonic)
		if err != nil {
			return err
		}

		coins := strings.Join(account.Coins, ",")
		if err := c.Commands().AddGenesisAccount(ctx, acc.Address, coins); err != nil {
			return err
		}
	}

	if err := c.configure(ctx); err != nil {
		return err
	}

	if _, err := c.plugin.Gentx(ctx, c.Commands(), Validator{
		Name:          conf.Validator.Name,
		StakingAmount: conf.Validator.Staked,
	}); err != nil {
		return err
	}

	if err := c.Commands().CollectGentxs(ctx); err != nil {
		return err
	}

	return c.start(ctx, conf)
}

func (c *Chain) start(ctx context.Context, conf conf.Config) error {
	g, ctx := errgroup.WithContext(ctx)

	// start the blockchain.
	g.Go(func() error { return c.plugin.Start(ctx, c.Commands(), conf) })

	// run relayer.
	go func() {
		if err := c.initRelayer(ctx, conf); err != nil && ctx.Err() == nil {
			fmt.Fprintf(c.stdLog(logStarport).err, "could not init relayer: %s", err)
		}
	}()

	// start the faucet if enabled.
	var (
		faucet, err     = c.Faucet(ctx)
		isFaucetEnabled = err != ErrFaucetIsNotEnabled
	)

	if isFaucetEnabled {
		if err == ErrFaucetAccountDoesNotExist {
			return &CannotBuildAppError{errors.Wrap(err, "faucet account doesn't exist")}
		}
		if err != nil {
			return err
		}

		g.Go(func() (err error) {
			if err := c.runFaucetServer(ctx, faucet); err != nil {
				return &CannotBuildAppError{err}
			}
			return nil
		})
	}

	// set the app as being served
	c.served = true

	// print the server addresses.
	fmt.Fprintf(c.stdLog(logStarport).out, "🌍 Running a Cosmos '%[1]v' app with Tendermint at %s.\n", c.app.Name, xurl.HTTP(conf.Servers.RPCAddr))
	fmt.Fprintf(c.stdLog(logStarport).out, "🌍 Running a server at %s (LCD)\n", xurl.HTTP(conf.Servers.APIAddr))

	if isFaucetEnabled {
		fmt.Fprintf(c.stdLog(logStarport).out, "🌍 Running a faucet at http://0.0.0.0:%d\n", conf.Faucet.Port)
	}

	fmt.Fprintf(c.stdLog(logStarport).out, "\n🚀 Get started: %s\n\n", xurl.HTTP(conf.Servers.DevUIAddr))

	return g.Wait()
}

func (c *Chain) watchAppFrontend(ctx context.Context) error {
	conf, err := c.Config()
	if err != nil {
		return err
	}
	vueFullPath := filepath.Join(c.app.Path, vuePath)
	if _, err := os.Stat(vueFullPath); os.IsNotExist(err) {
		return nil
	}
	frontendErr := &bytes.Buffer{}
	postExec := func(err error) error {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() > 0 {
			fmt.Fprintf(c.stdLog(logStarport).err, "%s\n%s",
				infoColor("skipping serving Vue frontend due to following errors:"), errorColor(frontendErr.String()))
		}
		return nil // ignore errors.
	}
	host, port, err := net.SplitHostPort(conf.Servers.FrontendAddr)
	if err != nil {
		return err
	}
	return cmdrunner.
		New(
			cmdrunner.DefaultWorkdir(vueFullPath),
			cmdrunner.DefaultStderr(frontendErr),
		).
		Run(ctx,
			step.New(
				step.Exec("npm", "i"),
				step.PostExec(postExec),
			),
			step.New(
				step.Exec("npm", "run", "serve"),
				step.Env(
					fmt.Sprintf("HOST=%s", host),
					fmt.Sprintf("PORT=%s", port),
					fmt.Sprintf("VUE_APP_API_COSMOS=%s", xurl.HTTP(conf.Servers.APIAddr)),
					fmt.Sprintf("VUE_APP_API_TENDERMINT=%s", xurl.HTTP(conf.Servers.RPCAddr)),
					fmt.Sprintf("VUE_APP_WS_TENDERMINT=%s/websocket", xurl.WS(conf.Servers.RPCAddr)),
				),
				step.PostExec(postExec),
			),
		)
}

func (c *Chain) runDevServer(ctx context.Context) error {
	config, err := c.Config()
	if err != nil {
		return err
	}

	grpcconn, grpcHandler, err := newGRPCWebProxyHandler(config.Servers.GRPCAddr)
	if err != nil {
		return err
	}
	defer grpcconn.Close()

	conf := Config{
		SdkVersion:      c.plugin.Name(),
		EngineAddr:      xurl.HTTP(config.Servers.RPCAddr),
		AppBackendAddr:  xurl.HTTP(config.Servers.APIAddr),
		AppFrontendAddr: xurl.HTTP(config.Servers.FrontendAddr),
	} // TODO get vals from const
	handler, err := newDevHandler(c.app, conf, grpcHandler)
	if err != nil {
		return err
	}

	return xhttp.Serve(ctx, &http.Server{
		Addr:    config.Servers.DevUIAddr,
		Handler: handler,
	})
}

func (c *Chain) runFaucetServer(ctx context.Context, faucet cosmosfaucet.Faucet) error {
	conf, err := c.Config()
	if err != nil {
		return err
	}

	return xhttp.Serve(ctx, &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", conf.Faucet.Port),
		Handler: faucet,
	})
}

// saveChainState runs the export command of the chain and store the exported genesis in the chain saved config
func (c *Chain) saveChainState(ctx context.Context) error {
	genesisPath, err := c.exportedGenesisPath()
	if err != nil {
		return err
	}

	return c.cmd.Export(ctx, genesisPath)
}

// importChainState imports the saved genesis in chain config to use it as the genesis
// nolint:unused
func (c *Chain) importChainState() error {
	exportGenesisPath, err := c.exportedGenesisPath()
	if err != nil {
		return err
	}
	genesisPath, err := c.GenesisPath()
	if err != nil {
		return err
	}

	return copy.Copy(exportGenesisPath, genesisPath)
}

// exportedGenesisPath gets the path of the exported genesis file
func (c *Chain) exportedGenesisPath() (string, error) {
	chainID, err := c.ID()
	if err != nil {
		return "", err
	}

	savePath := filepath.Join(chainSavePath, chainID)

	// ensure the path exists
	if err := os.MkdirAll(savePath, 0700); err != nil && !os.IsExist(err) {
		return "", err
	}

	return filepath.Join(savePath, exportedGenesis), nil
}

type CannotBuildAppError struct {
	Err error
}

func (e *CannotBuildAppError) Error() string {
	return fmt.Sprintf("cannot build app:\n\n\t%s", e.Err)
}

func (e *CannotBuildAppError) Unwrap() error {
	return e.Err
}
