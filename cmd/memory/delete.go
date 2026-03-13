package memory

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
)

func newDeleteCmd() *cobra.Command {
	var agentFlag string

	cmd := &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete a memory entry",
		Long:  "Delete a specific memory entry by key from an agent's memory namespace",
		Args:  cobra.ExactArgs(1),
		Example: `  # Delete a memory entry (from project directory)
  aphelion memory delete "last_request:+15551234567"

  # Delete for a specific agent
  aphelion memory delete "config:old_key" --agent review-management-agent`,
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
			endpoint := fmt.Sprintf("/v2/agents/%s/memory/%s", agentID, key)

			if err := client.Delete(endpoint); err != nil {
				return fmt.Errorf("failed to delete memory key %q: %w", key, err)
			}

			utils.PrintSuccess("Memory key %q deleted successfully", key)
			return nil
		},
	}

	cmd.Flags().StringVar(&agentFlag, "agent", "", "agent name or ID")

	return cmd
}
