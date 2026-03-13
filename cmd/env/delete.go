package env

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newDeleteCmd() *cobra.Command {
	var agentFlag string

	cmd := &cobra.Command{
		Use:   "delete <KEY>",
		Short: "Delete an environment variable from a deployed agent",
		Long:  "Remove an environment variable from a deployed agent's configuration.",
		Args:  cobra.ExactArgs(1),
		Example: `  # Delete an env var
  aphelion env delete TWILIO_PHONE_NUMBER

  # Delete for a specific agent
  aphelion env delete OLD_API_KEY --agent my-agent`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				return fmt.Errorf("session expired. Run: aphelion auth login")
			}

			key := args[0]

			agentID, err := resolveAgentID(agentFlag)
			if err != nil {
				return err
			}

			client := api.NewClient()
			endpoint := fmt.Sprintf("/v2/agents/%s/env/%s", agentID, key)

			if err := client.Delete(endpoint); err != nil {
				return fmt.Errorf("failed to delete environment variable %q: %w", key, err)
			}

			utils.PrintSuccess("Deleted environment variable %q", key)
			return nil
		},
	}

	cmd.Flags().StringVar(&agentFlag, "agent", "", "agent name or ID")

	return cmd
}
