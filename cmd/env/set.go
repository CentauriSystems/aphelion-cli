package env

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newSetCmd() *cobra.Command {
	var agentFlag string

	cmd := &cobra.Command{
		Use:   "set <KEY> <VALUE>",
		Short: "Set an environment variable for a deployed agent",
		Long:  "Set a secret environment variable that will be injected into the agent at runtime. Values are stored server-side and never displayed after being set.",
		Args:  cobra.ExactArgs(2),
		Example: `  # Set a Twilio phone number
  aphelion env set TWILIO_PHONE_NUMBER "+15551234567"

  # Set a SendGrid email
  aphelion env set SENDGRID_FROM_EMAIL "noreply@medgenie.com"

  # Set for a specific agent
  aphelion env set REVIEW_LINK "https://g.page/r/..." --agent my-agent`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				return fmt.Errorf("session expired. Run: aphelion auth login")
			}

			key := args[0]
			value := args[1]

			agentID, err := resolveAgentID(agentFlag)
			if err != nil {
				return err
			}

			client := api.NewClient()
			endpoint := fmt.Sprintf("/v2/agents/%s/env/%s", agentID, key)

			body := map[string]string{
				"value": value,
			}

			if err := client.Put(endpoint, body, nil); err != nil {
				return fmt.Errorf("failed to set environment variable %q: %w", key, err)
			}

			name := resolveAgentName()
			utils.PrintSuccess("Set %s for agent %s", key, name)
			return nil
		},
	}

	cmd.Flags().StringVar(&agentFlag, "agent", "", "agent name or ID")

	return cmd
}
