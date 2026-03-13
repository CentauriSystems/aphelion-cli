package schedule

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newDisableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable <agent-name>",
		Short: "Disable the schedule for a deployed agent",
		Long:  "Disable the cron schedule for a deployed agent without removing it.",
		Example: `  aphelion schedule disable review-management-agent`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			agentName := args[0]

			s := utils.NewSpinner("Disabling schedule...")
			s.Start()

			client := api.NewClient()

			body := map[string]bool{
				"enabled": false,
			}

			var result map[string]interface{}
			err := client.Patch(fmt.Sprintf("/v2/agents/%s/schedule", agentName), body, &result)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to disable schedule: %v", err)
				return err
			}

			utils.PrintSuccess("Schedule disabled for %s", agentName)
			return nil
		},
	}

	return cmd
}
