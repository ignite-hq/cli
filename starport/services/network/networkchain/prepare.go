package networkchain

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	launchtypes "github.com/tendermint/spn/x/launch/types"
	"github.com/tendermint/starport/starport/pkg/cosmosaddress"
	"github.com/tendermint/starport/starport/pkg/cosmosutil"
	"github.com/tendermint/starport/starport/pkg/events"
)

// Prepare queries launch information and prepare the chain to be launched from these information
func (c Chain) Prepare(ctx context.Context) error {
	// chain initialization
	chainHome, err := b.chain.Home()
	if err != nil {
		return err
	}

	_, err = os.Stat(chainHome)

	// nolint:gocritic
	if os.IsNotExist(err) {
		// if no config exists, we perform a full initialization of the chain with a new validator key
		if err := b.Init(ctx); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		// if config and validator key already exists we only build the chain and initialized the genesis
		b.builder.ev.Send(events.New(events.StatusOngoing, "Building the blockchain"))
		if _, err := b.chain.Build(ctx, ""); err != nil {
			return err
		}
		b.builder.ev.Send(events.New(events.StatusDone, "Blockchain built"))

		b.builder.ev.Send(events.New(events.StatusOngoing, "Initializing the genesis"))
		if err := b.initGenesis(ctx); err != nil {
			return err
		}
		b.builder.ev.Send(events.New(events.StatusDone, "Genesis initialized"))
	}

	return b.buildGenesis(ctx)
}

// buildGenesis builds the genesis for the chain from the launch approved requests
func (c Chain) buildGenesis(ctx context.Context) error {
	b.builder.ev.Send(events.New(events.StatusOngoing, "Building the genesis"))

	// get the genesis accounts and apply them to the genesis
	genesisAccounts, err := b.builder.GenesisAccounts(ctx, b.launchID)
	if err != nil {
		return errors.Wrap(err, "error querying genesis accounts")
	}
	if err := b.applyGenesisAccounts(ctx, genesisAccounts); err != nil {
		return errors.Wrap(err, "error applying genesis accounts to genesis")
	}

	// get the genesis vesting accounts and apply them to the genesis
	vestingAccounts, err := b.builder.VestingAccounts(ctx, b.launchID)
	if err != nil {
		return errors.Wrap(err, "error querying vesting accounts")
	}
	if err := b.applyVestingAccounts(ctx, vestingAccounts); err != nil {
		return errors.Wrap(err, "error applying vesting accounts to genesis")
	}

	// get the genesis validators, gather gentxs and modify config to include the peers
	genesisValidators, err := b.builder.GenesisValidators(ctx, b.launchID)
	if err != nil {
		return errors.Wrap(err, "error querying genesis validators")
	}
	if err := b.applyGenesisValidators(ctx, genesisValidators); err != nil {
		return errors.Wrap(err, "error applying genesis validators to genesis")
	}

	// set the genesis time for the chain
	genesisPath, err := b.chain.GenesisPath()
	if err != nil {
		return errors.Wrap(err, "genesis of the blockchain can't be read")
	}
	if err := cosmosutil.SetGenesisTime(genesisPath, b.launchTime); err != nil {
		return errors.Wrap(err, "genesis time can't be set")
	}

	b.builder.ev.Send(events.New(events.StatusDone, "Genesis built"))

	return nil
}

// applyGenesisAccounts adds the genesis account into the genesis using the chain CLI
func (c Chain) applyGenesisAccounts(ctx context.Context, genesisAccs []launchtypes.GenesisAccount) error {
	var err error
	// TODO: detect the correct prefix
	prefix := "cosmos"

	cmd, err := b.chain.Commands(ctx)
	if err != nil {
		return err
	}

	for _, acc := range genesisAccs {
		// change the address prefix to the target chain prefix
		acc.Address, err = cosmosaddress.ChangePrefix(acc.Address, prefix)
		if err != nil {
			return err
		}

		// call add genesis account cli command
		err = cmd.AddGenesisAccount(ctx, acc.Address, acc.Coins.String())
		if err != nil {
			return err
		}
	}

	return nil
}

// applyVestingAccounts adds the genesis vesting account into the genesis using the chain CLI
func (c Chain) applyVestingAccounts(ctx context.Context, vestingAccs []launchtypes.VestingAccount) error {
	var err error
	prefix := "cosmos"

	cmd, err := b.chain.Commands(ctx)
	if err != nil {
		return err
	}

	for _, acc := range vestingAccs {
		acc.Address, err = cosmosaddress.ChangePrefix(acc.Address, prefix)
		if err != nil {
			return err
		}

		// only delayed vesting option is supported for now
		delayedVesting := acc.VestingOptions.GetDelayedVesting()
		if delayedVesting == nil {
			return fmt.Errorf("invalid vesting option for account %s", acc.Address)
		}

		// call add genesis account cli command with delayed vesting option
		err = cmd.AddVestingAccount(
			ctx,
			acc.Address,
			acc.StartingBalance.String(),
			delayedVesting.Vesting.String(),
			delayedVesting.EndTime,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// applyGenesisValidators gathers the validator gentxs into the genesis and add peers in config
func (c Chain) applyGenesisValidators(ctx context.Context, genesisVals []launchtypes.GenesisValidator) error {
	// no validator
	if len(genesisVals) == 0 {
		return nil
	}

	// reset the gentx directory
	gentxDir, err := b.chain.GentxsPath()
	if err != nil {
		return err
	}
	if err := os.RemoveAll(gentxDir); err != nil {
		return err
	}
	if err := os.MkdirAll(gentxDir, 0700); err != nil {
		return err
	}

	// write gentxs
	for i, val := range genesisVals {
		gentxPath := filepath.Join(gentxDir, fmt.Sprintf("gentx%d.json", i))
		if err = ioutil.WriteFile(gentxPath, val.GenTx, 0666); err != nil {
			return err
		}
	}

	// gather gentxs
	cmd, err := b.chain.Commands(ctx)
	if err != nil {
		return err
	}
	if err := cmd.CollectGentxs(ctx); err != nil {
		return err
	}

	return b.updateConfigFromGenesisValidators(genesisVals)
}

// updateConfigFromGenesisValidators adds the peer addresses into the config.toml of the chain
func (bc Chain) updateConfigFromGenesisValidators(genesisVals []launchtypes.GenesisValidator) error {
	var p2pAddresses []string
	for _, val := range genesisVals {
		p2pAddresses = append(p2pAddresses, val.Peer)
	}

	// set persistent peers
	configPath, err := b.chain.ConfigTOMLPath()
	if err != nil {
		return err
	}
	configToml, err := toml.LoadFile(configPath)
	if err != nil {
		return err
	}
	configToml.Set("p2p.persistent_peers", strings.Join(p2pAddresses, ","))
	if err != nil {
		return err
	}

	// save config.toml file
	configTomlFile, err := os.OpenFile(configPath, os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer configTomlFile.Close()
	_, err = configToml.WriteTo(configTomlFile)
	return err
}
