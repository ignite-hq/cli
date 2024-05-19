package ignitecmd

import (
	"github.com/ignite/cli/v29/ignite/pkg/cliui"
	"github.com/ignite/cli/v29/ignite/services/doctor"
	"github.com/spf13/cobra"
)

func NewDoctor() *cobra.Command {
	return &cobra.Command{
		Use:    "doctor",
		Short:  "Fix chain configuration",
		Hidden: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			session := cliui.New()
			defer session.End()

			doc := doctor.New(doctor.CollectEvents(session.EventBus()))

			if err := doc.MigrateConfig(cmd.Context()); err != nil {
				return err
			}

			if err := doc.MigratePluginsConfig(); err != nil {
				return err
			}

			return doc.FixDependencyTools(cmd.Context())
		},
	}
}
