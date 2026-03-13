package memory

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newGetCmd() *cobra.Command {
	var (
		agentFlag string
		fromFlag  string
	)

	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a memory entry by key",
		Long:  "Retrieve a specific memory entry by key from an agent's memory namespace",
		Args:  cobra.ExactArgs(1),
		Example: `  # Get a memory entry (from project directory)
  aphelion memory get "last_request:+15551234567"

  # Get a memory entry for a specific agent
  aphelion memory get "config:review_link" --agent review-management-agent

  # Cross-agent memory read (requires permission)
  aphelion memory get "research:topic" --from researcher-agent`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				return fmt.Errorf("session expired. Run: aphelion auth login")
			}

			key := args[0]

			// Determine which agent's memory to read from
			var targetAgentID string
			if fromFlag != "" {
				// Cross-agent read
				targetAgentID = fromFlag
			} else {
				agentID, err := resolveAgentID(agentFlag)
				if err != nil {
					return err
				}
				targetAgentID = agentID
			}

			client := api.NewClient()
			endpoint := fmt.Sprintf("/v2/agents/%s/memory/%s", targetAgentID, key)

			var result interface{}
			if err := client.Get(endpoint, &result); err != nil {
				return fmt.Errorf("failed to get memory key %q: %w", key, err)
			}

			// Print as formatted JSON
			output, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format response: %w", err)
			}

			fmt.Println(string(output))
			return nil
		},
	}

	cmd.Flags().StringVar(&agentFlag, "agent", "", "agent name or ID")
	cmd.Flags().StringVar(&fromFlag, "from", "", "read from another agent's memory (requires permission)")

	return cmd
}
