package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
)

func newOpenCmd() *cobra.Command {
	targets := map[string]string{
		"":            "https://beta.console.aphl.ai",
		"agents":      "https://beta.console.aphl.ai/agents",
		"marketplace": "https://beta.console.aphl.ai/marketplace",
		"history":     "https://beta.console.aphl.ai/history",
		"docs":        "https://api.aphl.ai/docs",
	}

	cmd := &cobra.Command{
		Use:   "open [target]",
		Short: "Open Aphelion console in your browser",
		Long:  "Open the Aphelion console or a specific page in your default browser.",
		Example: `  # Open the console dashboard
  aphelion open

  # Open the agents page
  aphelion open agents

  # Open the marketplace
  aphelion open marketplace

  # Open documentation
  aphelion open docs`,
		ValidArgs: []string{"agents", "marketplace", "history", "docs"},
		Args:      cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := ""
			if len(args) > 0 {
				target = args[0]
			}

			url, ok := targets[target]
			if !ok {
				return fmt.Errorf("Unknown target %q. Valid targets: agents, marketplace, history, docs", target)
			}

			utils.OpenBrowserWithFallback(url)
			return nil
		},
	}

	return cmd
}
