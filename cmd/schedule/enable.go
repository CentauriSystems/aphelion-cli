package schedule

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newEnableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable <agent-name>",
		Short: "Enable the schedule for a deployed agent",
		Long:  "Enable a previously configured cron schedule for a deployed agent.",
		Example: `  aphelion schedule enable review-management-agent`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			agentName := args[0]

			s := utils.NewSpinner("Enabling schedule...")
			s.Start()

			client := api.NewClient()

			body := map[string]bool{
				"enabled": true,
			}

			var result map[string]interface{}
			err := client.Patch(fmt.Sprintf("/v2/agents/%s/schedule", agentName), body, &result)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to enable schedule: %v", err)
				return err
			}

			utils.PrintSuccess("Schedule enabled for %s", agentName)
			return nil
		},
	}

	return cmd
}
