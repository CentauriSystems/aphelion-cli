package schedule

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newSetCmd() *cobra.Command {
	var cronExpr string

	cmd := &cobra.Command{
		Use:   "set <agent-name>",
		Short: "Set a cron schedule for a deployed agent",
		Long:  "Set a cron schedule for a deployed agent. The agent will run automatically at the specified times.",
		Example: `  aphelion schedule set review-management-agent --cron "0 9 * * MON-FRI"
  aphelion schedule set data-pipeline --cron "*/30 * * * *"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			agentName := args[0]

			s := utils.NewSpinner("Setting schedule...")
			s.Start()

			client := api.NewClient()

			body := map[string]string{
				"cron": cronExpr,
			}

			var result map[string]interface{}
			err := client.Post(fmt.Sprintf("/v2/agents/%s/schedule", agentName), body, &result)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to set schedule: %v", err)
				return err
			}

			humanReadable := cronExpr
			if desc, ok := result["description"].(string); ok && desc != "" {
				humanReadable = desc
			}

			utils.PrintSuccess("Schedule set for %s", agentName)
			fmt.Printf("  Cron:     %s\n", cronExpr)
			fmt.Printf("  Runs at:  %s\n", humanReadable)

			return nil
		},
	}

	cmd.Flags().StringVar(&cronExpr, "cron", "", "Cron expression (required)")
	cmd.MarkFlagRequired("cron")

	return cmd
}
