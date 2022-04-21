package ignitecmd

import (
	"fmt"
	"path/filepath"

	"github.com/ignite-hq/cli/ignite/pkg/events"
	"github.com/ignite-hq/cli/ignite/pkg/goenv"
	"github.com/ignite-hq/cli/ignite/services/network"
	"github.com/ignite-hq/cli/ignite/services/network/networkchain"
	"github.com/spf13/cobra"
)

const (
	flagForce = "force"
)

// NewNetworkChainPrepare returns a new command to prepare the chain for launch
func NewNetworkChainPrepare() *cobra.Command {
	c := &cobra.Command{
		Use:   "prepare [launch-id]",
		Short: "Prepare the chain for launch",
		Args:  cobra.ExactArgs(1),
		RunE:  networkChainPrepareHandler,
	}

	c.Flags().BoolP(flagForce, "f", false, "Force the prepare command to run even if the chain is not launched")
	c.Flags().AddFlagSet(flagNetworkFrom())
	c.Flags().AddFlagSet(flagSetKeyringBackend())
	c.Flags().AddFlagSet(flagSetHome())

	return c
}

func networkChainPrepareHandler(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool(flagForce)

	nb, err := newNetworkBuilder(cmd)
	if err != nil {
		return err
	}
	defer nb.Cleanup()

	// parse launch ID
	launchID, err := network.ParseID(args[0])
	if err != nil {
		return err
	}

	n, err := nb.Network()
	if err != nil {
		return err
	}

	// fetch chain information
	chainLaunch, err := n.ChainLaunch(cmd.Context(), launchID)
	if err != nil {
		return err
	}

	if !force && !chainLaunch.LaunchTriggered {
		return fmt.Errorf("chain %d has not launched yet. use --force to prepare anyway", launchID)
	}

	c, err := nb.Chain(networkchain.SourceLaunch(chainLaunch))
	if err != nil {
		return err
	}

	// fetch the information to construct genesis
	genesisInformation, err := n.GenesisInformation(cmd.Context(), launchID)
	if err != nil {
		return err
	}

	if err := c.Prepare(cmd.Context(), genesisInformation); err != nil {
		return err
	}

	chainHome, err := c.Home()
	if err != nil {
		return err
	}
	binaryName, err := c.BinaryName()
	if err != nil {
		return err
	}
	binaryDir := filepath.Dir(filepath.Join(goenv.Bin(), binaryName))

	nb.EventBus.Send(events.New(events.StatusDone, fmt.Sprintf("Chain is prepared for launch")))
	commandStr := fmt.Sprintf("%s start --home %s", binaryName, chainHome)
	nb.EventBus.Send(
		events.New(
			events.StatusInfo,
			fmt.Sprintf(
				"You can start your node by running the following command:\n\t%s/%s",
				binaryDir,
				infoColor(commandStr),
			),
		),
	)
	return nil
}
