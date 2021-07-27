package starportcmd

import (
	"github.com/spf13/cobra"
	"os"
)

func Completion() *cobra.Command {

	// completionCmd represents the completion command
	c := &cobra.Command{
		Use:   "completion",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:

  $ source <(starport completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ starport completion bash > /etc/bash_completion.d/starport
  # macOS:
  $ starport completion bash > /usr/local/etc/bash_completion.d/starport

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ starport completion zsh > "${fpath[1]}/_starport"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ starport completion fish | source

  # To load completions for each session, execute once:
  $ starport completion fish > ~/.config/fish/completions/starport.fish

PowerShell:

  PS> starport completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> starport completion powershell > starport.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}
	return c
}
