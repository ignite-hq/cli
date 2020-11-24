package starportcmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/tendermint/starport/starport/pkg/clispinner"
	"github.com/tendermint/starport/starport/pkg/events"
	"github.com/tendermint/starport/starport/services/chain"
)

var (
	infoColor = color.New(color.FgYellow).SprintFunc()
)

// New creates a new root command for `starport` with its sub commands.
func New() *cobra.Command {
	c := &cobra.Command{
		Use:           "starport",
		Short:         "A tool for scaffolding out Cosmos applications",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	c.AddCommand(NewApp())
	c.AddCommand(NewType())
	c.AddCommand(NewServe())
	c.AddCommand(NewBuild())
	c.AddCommand(NewModule())
	c.AddCommand(NewRelayer())
	c.AddCommand(NewVersion())
	c.AddCommand(NewNetwork())
	c.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	return c
}

func getModule(path string) string {
	goModFile, err := ioutil.ReadFile(filepath.Join(path, "go.mod"))
	if err != nil {
		log.Fatal(err)
	}
	moduleString := strings.Split(string(goModFile), "\n")[0]
	modulePath := strings.ReplaceAll(moduleString, "module ", "")
	return modulePath
}

func logLevel(cmd *cobra.Command) chain.LogLevel {
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		return chain.LogVerbose
	}
	return chain.LogRegular
}

func printEvents(bus events.Bus, s *clispinner.Spinner) {
	for event := range bus {
		if event.IsOngoing() {
			s.SetText(event.Text())
			s.Start()
		} else {
			s.Stop()
			fmt.Printf("%s %s\n", color.New(color.FgGreen).SprintFunc()("✔"), event.Description)
		}
	}
}
