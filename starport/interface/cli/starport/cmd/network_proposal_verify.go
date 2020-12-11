package starportcmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/tendermint/starport/starport/pkg/clispinner"
	"github.com/tendermint/starport/starport/pkg/numbers"
	"io"
	"io/ioutil"
	"os"
)

const (
	flagDebug = "debug"
)

func NewNetworkProposalVerify() *cobra.Command {
	c := &cobra.Command{
		Use:   "verify [chain-id] [number<,...>]",
		Short: "Simulate and verify proposals validity",
		RunE:  networkProposalVerifyHandler,
		Args:  cobra.ExactArgs(2),
	}
	c.Flags().Bool(flagDebug, false, "show output of verification command in the console")
	return c
}

func networkProposalVerifyHandler(cmd *cobra.Command, args []string) error {
	s := clispinner.New()
	defer s.Stop()

	var (
		chainID			= args[0]
		proposalList 	= args[1]
	)

	nb, err := newNetworkBuilder()
	if err != nil {
		return err
	}

	ids, err := numbers.ParseList(proposalList)
	if err != nil {
		return err
	}

	s.SetText("Verifying proposals...")
	s.Start()

	// Check verbose flag
	var out io.Writer
	debugSet, err := cmd.Flags().GetBool(flagDebug)
	if err != nil {
		return err
	}
	if debugSet {
		out = os.Stdout
	} else {
		out = ioutil.Discard
	}

	verified, err := nb.VerifyProposals(cmd.Context(), chainID, ids, out)
	if err != nil {
		return err
	}
	if verified {
		fmt.Printf("Proposal(s) %s verified 🔍✅️\n", numbers.List(ids, "#"))
	} else {
		fmt.Printf("Genesis from proposal(s) %s is invalid 🔍❌️\n", numbers.List(ids, "#"))
	}

	return nil
}