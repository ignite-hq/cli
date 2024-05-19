package ignitecmd

import (
	"github.com/ignite/cli/v29/ignite/pkg/cosmosaccount"
	"github.com/spf13/cobra"
)

func NewAccountList() *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "Show a list of all accounts",
		RunE:  accountListHandler,
	}

	c.Flags().AddFlagSet(flagSetAccountPrefixes())

	return c
}

func accountListHandler(cmd *cobra.Command, _ []string) error {
	ca, err := cosmosaccount.New(
		cosmosaccount.WithKeyringBackend(getKeyringBackend(cmd)),
		cosmosaccount.WithHome(getKeyringDir(cmd)),
	)
	if err != nil {
		return err
	}

	accounts, err := ca.List()
	if err != nil {
		return err
	}

	return printAccounts(cmd, accounts...)
}
