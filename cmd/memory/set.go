package memory

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
)

func newSetCmd() *cobra.Command {
	var (
		agentFlag string
		ttlFlag   string
	)

	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a memory entry",
		Long:  "Write a key-value pair to an agent's memory. Value must be a valid JSON string.",
		Args:  cobra.ExactArgs(2),
		Example: `  # Set a simple value
  aphelion memory set "config:review_link" '"https://g.page/r/..."'

  # Set a JSON object with TTL
  aphelion memory set "last_request:+15551234567" '{"patient": "Jane", "channel": "sms"}' --ttl 7d

  # Set for a specific agent
  aphelion memory set "config:key" '"value"' --agent review-management-agent`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				return fmt.Errorf("session expired. Run: aphelion auth login")
			}

			key := args[0]
			valueStr := args[1]

			agentID, err := resolveAgentID(agentFlag)
			if err != nil {
				return err
			}

			// Parse the value as JSON
			var parsedValue interface{}
			if err := json.Unmarshal([]byte(valueStr), &parsedValue); err != nil {
				return fmt.Errorf("value must be valid JSON: %w\nExample: aphelion memory set mykey '{\"foo\": \"bar\"}'", err)
			}

			// Build request body
			body := map[string]interface{}{
				"value": parsedValue,
			}
			if ttlFlag != "" {
				body["ttl"] = ttlFlag
			}

			client := api.NewClient()
			endpoint := fmt.Sprintf("/v2/agents/%s/memory/%s", agentID, key)

			if err := client.Put(endpoint, body, nil); err != nil {
				return fmt.Errorf("failed to set memory key %q: %w", key, err)
			}

			utils.PrintSuccess("Memory key %q set successfully", key)
			return nil
		},
	}

	cmd.Flags().StringVar(&agentFlag, "agent", "", "agent name or ID")
	cmd.Flags().StringVar(&ttlFlag, "ttl", "", "time to live (e.g. 7d, 24h, 30m)")

	return cmd
}
