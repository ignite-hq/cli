package networkbuilder

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
	"golang.org/x/sync/errgroup"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/tendermint/starport/starport/pkg/availableport"
	"github.com/tendermint/starport/starport/pkg/cmdrunner"
	"github.com/tendermint/starport/starport/pkg/cmdrunner/step"
	"github.com/tendermint/starport/starport/pkg/confile"
	"github.com/tendermint/starport/starport/pkg/cosmosver"
	"github.com/tendermint/starport/starport/pkg/events"
	"github.com/tendermint/starport/starport/pkg/gomodulepath"
	"github.com/tendermint/starport/starport/pkg/spn"
	"github.com/tendermint/starport/starport/pkg/xchisel"
	"github.com/tendermint/starport/starport/services/chain"
)

// Builder is network builder.
type Builder struct {
	ev        events.Bus
	spnclient *spn.Client
}

type Option func(*Builder)

// CollectEvents collects events from Builder.
func CollectEvents(ev events.Bus) Option {
	return func(b *Builder) {
		b.ev = ev
	}
}

// New creates a Builder.
func New(spnclient *spn.Client, options ...Option) (*Builder, error) {
	b := &Builder{
		spnclient: spnclient,
	}
	for _, opt := range options {
		opt(b)
	}
	return b, nil
}

// InitBlockchainFromChainID initializes blockchain from chain id.
func (b *Builder) InitBlockchainFromChainID(ctx context.Context, chainID string, mustNotInitializedBefore bool) (*Blockchain, error) {
	account, err := b.AccountInUse()
	if err != nil {
		return nil, err
	}
	chain, err := b.spnclient.ShowChain(ctx, account.Name, chainID)
	if err != nil {
		return nil, err
	}
	return b.InitBlockchainFromURL(ctx, chainID, chain.URL, chain.Hash, mustNotInitializedBefore)
}

// InitBlockchainFromURL initializes blockchain from a remote git repo.
func (b *Builder) InitBlockchainFromURL(ctx context.Context, chainID, url, rev string, mustNotInitializedBefore bool) (*Blockchain, error) {
	appPath, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}

	b.ev.Send(events.New(events.StatusOngoing, "Pulling the blockchain"))

	// clone the repo.
	repo, err := git.PlainCloneContext(ctx, appPath, false, &git.CloneOptions{
		URL: url,
	})
	if err != nil {
		return nil, err
	}

	var hash plumbing.Hash

	// checkout to the revision if provided, otherwise default branch is used.
	if rev != "" {
		wt, err := repo.Worktree()
		if err != nil {
			return nil, err
		}
		h, err := repo.ResolveRevision(plumbing.Revision(rev))
		if err != nil {
			return nil, err
		}
		hash = *h
		wt.Checkout(&git.CheckoutOptions{
			Hash: hash,
		})
	} else {
		ref, err := repo.Head()
		if err != nil {
			return nil, err
		}
		hash = ref.Hash()
	}

	b.ev.Send(events.New(events.StatusDone, "Pulled the blockchain"))

	return newBlockchain(ctx, b, chainID, appPath, url, hash.String(), mustNotInitializedBefore)
}

// InitBlockchainFromPath initializes blockchain from a local git repo.
//
// It uses the HEAD(latest commit in currently checked out branch) as the source code of blockchain.
//
// TODO: It requires that there will be no unstaged changes in the code and HEAD is synced with the upstream
// branch (if there is one).
func (b *Builder) InitBlockchainFromPath(ctx context.Context, chainID string, appPath string,
	mustNotInitializedBefore bool) (*Blockchain, error) {
	repo, err := git.PlainOpen(appPath)
	if err != nil {
		return nil, err
	}

	// check if there are un-committed changes.
	wt, err := repo.Worktree()
	if err != nil {
		return nil, err
	}
	status, err := wt.Status()
	if err != nil {
		return nil, err
	}
	if !status.IsClean() {
		return nil, errors.New("please either revert or commit your changes")
	}

	// find out remote's url.
	// TODO use the associated upstream branch's remote.
	remotes, err := repo.Remotes()
	if err != nil {
		return nil, err
	}
	if len(remotes) == 0 {
		return nil, errors.New("please push your blockchain first")
	}
	remote := remotes[0]
	rc := remote.Config()
	if len(rc.URLs) == 0 {
		return nil, errors.New("cannot find remote's url")
	}
	url := rc.URLs[0]

	// find the hash pointing to HEAD.
	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}
	hash := ref.Hash()
	if err != nil {
		return nil, err
	}

	return newBlockchain(ctx, b, chainID, appPath, url, hash.String(), mustNotInitializedBefore)
}

// StartChain downloads the final version version of Genesis on the first start or fails if Genesis
// has not finalized yet.
// After overwriting the downloaded Genesis on top of app's home dir, it starts blockchain by
// executing the start command on its appd binary with optionally provided flags.
func (b *Builder) StartChain(ctx context.Context, chainID string, flags []string) error {
	c, err := b.ShowChain(ctx, chainID)
	if err != nil {
		return err
	}

	info, err := b.LaunchInformation(ctx, chainID)
	if err != nil {
		return err
	}

	// find out the app's name form url.
	u, err := url.Parse(c.URL)
	if err != nil {
		return err
	}
	importPath := path.Join(u.Host, u.Path)
	path, err := gomodulepath.Parse(importPath)
	if err != nil {
		return err
	}

	app := chain.App{
		ChainID: chainID,
		Name:    path.Root,
		Version: cosmosver.Stargate,
	}
	ch, err := chain.New(app, true, chain.LogSilent)
	if err != nil {
		return err
	}

	if len(info.GenTxs) == 0 {
		return errors.New("There are no approved validators yet")
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// overwrite genesis with initial genesis.
	appHome := filepath.Join(homedir, app.ND())
	os.Rename(initialGenesisPath(appHome), genesisPath(appHome))

	// make sure that Genesis' genesis_time is set to chain's creation time on SPN.
	cf := confile.New(confile.DefaultJSONEncodingCreator, genesisPath(appHome))
	var genesis map[string]interface{}
	if err := cf.Load(&genesis); err != nil {
		return err
	}
	genesis["genesis_time"] = c.CreatedAt.UTC().Format(time.RFC3339)
	if err := cf.Save(genesis); err != nil {
		return err
	}

	// add the genesis accounts
	for _, account := range info.GenesisAccounts {
		if err = ch.AddGenesisAccount(ctx, chain.Account{
			Address: account.Address.String(),
			Coins:   account.Coins.String(),
		}); err != nil {
			return err
		}
	}

	// reset gentx directory
	dir, err := ioutil.ReadDir(filepath.Join(appHome, "config/gentx"))
	if err != nil {
		return err
	}
	for _, d := range dir {
		if err := os.RemoveAll(filepath.Join(appHome, "config/gentx", d.Name())); err != nil {
			return err
		}
	}

	// add and collect the gentxs
	for i, gentx := range info.GenTxs {
		// Save the gentx in the gentx directory
		gentxPath := filepath.Join(appHome, fmt.Sprintf("config/gentx/gentx%v.json", i))
		if err = ioutil.WriteFile(gentxPath, gentx, 0666); err != nil {
			return err
		}
	}
	if err = ch.CollectGentx(ctx); err != nil {
		return err
	}

	// prep peer configs.
	p2pAddresses := info.Peers
	chiselAddreses := make(map[string]int) // server addr-local p2p port pair.
	ports, err := availableport.Find(len(info.Peers))
	if err != nil {
		return err
	}
	time.Sleep(time.Second * 2) // make sure that ports are released by the OS before being used.

	if xchisel.IsEnabled() {
		for i, peer := range info.Peers {
			localPort := ports[i]
			sp := strings.Split(peer, "@")
			nodeID := sp[0]
			serverAddr := sp[1]

			p2pAddresses[i] = fmt.Sprintf("%s@127.0.0.1:%d", nodeID, localPort)
			chiselAddreses[serverAddr] = localPort
		}
	}

	// save the finalized version of config.toml with peers.
	configTomlPath := filepath.Join(appHome, "config/config.toml")
	configToml, err := toml.LoadFile(configTomlPath)
	if err != nil {
		return err
	}
	configToml.Set("p2p.persistent_peers", strings.Join(p2pAddresses, ","))
	configToml.Set("p2p.allow_duplicate_ip", true)
	configTomlFile, err := os.OpenFile(configTomlPath, os.O_RDWR|os.O_TRUNC, 644)
	if err != nil {
		return err
	}
	defer configTomlFile.Close()
	if _, err = configToml.WriteTo(configTomlFile); err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)

	// run the start command of the chain.
	g.Go(func() error {
		return cmdrunner.New().Run(ctx, step.New(
			step.Exec(
				app.D(),
				append([]string{"start"}, flags...)...,
			),
			step.Stdout(os.Stdout),
			step.Stderr(os.Stderr),
		))
	})

	if xchisel.IsEnabled() {
		// start Chisel server.
		g.Go(func() error {
			return xchisel.StartServer(ctx, xchisel.DefaultServerPort)
		})

		// start Chisel clients for all other validators.
		for serverAddr, localPort := range chiselAddreses {
			serverAddr, localPort := serverAddr, localPort
			g.Go(func() error {
				return xchisel.StartClient(ctx, serverAddr, fmt.Sprintf("%d", localPort), "26656")
			})
		}
	}

	return g.Wait()
}
