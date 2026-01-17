package main

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

  $ source <(configdiff completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ configdiff completion bash > /etc/bash_completion.d/configdiff
  # macOS:
  $ configdiff completion bash > $(brew --prefix)/etc/bash_completion.d/configdiff

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ configdiff completion zsh > "${fpath[1]}/_configdiff"

  # You will need to start a new shell for this setup to take effect.

Fish:

  $ configdiff completion fish | source

  # To load completions for each session, execute once:
  $ configdiff completion fish > ~/.config/fish/completions/configdiff.fish

PowerShell:

  PS> configdiff completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> configdiff completion powershell > configdiff.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		switch args[0] {
		case "bash":
			err = cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			err = cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			err = cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			err = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return err
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
