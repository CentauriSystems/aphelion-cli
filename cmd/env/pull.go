package env

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newPullCmd() *cobra.Command {
	var agentFlag string

	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull environment variable keys from deployed agent to local .env",
		Long: `Pull environment variable keys from a deployed agent and write them to a local .env file.
Values are masked since they cannot be retrieved from the server.`,
		Example: `  # Pull env vars to .env
  aphelion env pull

  # Pull for a specific agent
  aphelion env pull --agent review-management-agent`,
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
				return fmt.Errorf("failed to pull environment variables: %w", err)
			}

			if len(result.Keys) == 0 {
				utils.PrintInfo("No environment variables configured on deployed agent")
				return nil
			}

			// Write .env file with masked values
			var content string
			content += "# Environment variables pulled from deployed agent\n"
			content += "# Values are masked — fill in with actual values\n\n"
			for _, key := range result.Keys {
				content += fmt.Sprintf("%s=********\n", key)
			}

			if err := os.WriteFile(".env", []byte(content), 0600); err != nil {
				return fmt.Errorf("failed to write .env file: %w", err)
			}

			utils.PrintSuccess("Pulled %d environment variables to .env (values masked)", len(result.Keys))
			return nil
		},
	}

	cmd.Flags().StringVar(&agentFlag, "agent", "", "agent name or ID")

	return cmd
}
