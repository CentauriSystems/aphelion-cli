package env

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newListCmd() *cobra.Command {
	var agentFlag string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List environment variable keys for a deployed agent",
		Long:  "List all environment variable keys configured for a deployed agent. Values are never shown.",
		Example: `  # List env vars for current project agent
  aphelion env list

  # List for a specific agent
  aphelion env list --agent review-management-agent`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				return fmt.Errorf("session expired. Run: aphelion auth login")
			}

			agentID, err := resolveAgentID(agentFlag)
			if err != nil {
				return err
			}

			client := api.NewClient()
			endpoint := fmt.Sprintf("/v2/agents/%s/env", agentID)

			var result struct {
				Keys []string `json:"keys"`
			}
			if err := client.Get(endpoint, &result); err != nil {
				return fmt.Errorf("failed to list environment variables: %w", err)
			}

			if len(result.Keys) == 0 {
				utils.PrintInfo("No environment variables configured")
				return nil
			}

			utils.PrintInfo("Environment variables (%d):", len(result.Keys))
			for _, key := range result.Keys {
				fmt.Printf("  %s\n", key)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&agentFlag, "agent", "", "agent name or ID")

	return cmd
}
