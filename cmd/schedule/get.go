package schedule

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <agent-name>",
		Short: "Get the schedule for a deployed agent",
		Long:  "Display the current cron schedule, enabled status, and next run times for a deployed agent.",
		Example: `  aphelion schedule get review-management-agent
  aphelion schedule get data-pipeline`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			agentName := args[0]

			s := utils.NewSpinner("Fetching schedule...")
			s.Start()

			client := api.NewClient()

			var result map[string]interface{}
			err := client.Get(fmt.Sprintf("/v2/agents/%s/schedule", agentName), &result)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to get schedule: %v", err)
				return err
			}

			bold := color.New(color.Bold)
			bold.Printf("Schedule for %s\n\n", agentName)

			cronExpr, _ := result["cron"].(string)
			if cronExpr == "" {
				utils.PrintInfo("No schedule configured.")
				fmt.Println()
				fmt.Println("Set one with:")
				fmt.Printf("  aphelion schedule set %s --cron \"0 9 * * MON-FRI\"\n", agentName)
				return nil
			}

			enabled, _ := result["enabled"].(bool)
			enabledStr := color.GreenString("enabled")
			if !enabled {
				enabledStr = color.YellowString("disabled")
			}

			fmt.Printf("  Cron:     %s\n", cronExpr)
			fmt.Printf("  Status:   %s\n", enabledStr)

			if nextRuns, ok := result["next_runs"].([]interface{}); ok && len(nextRuns) > 0 {
				fmt.Println()
				bold.Println("  Next runs:")
				limit := 5
				if len(nextRuns) < limit {
					limit = len(nextRuns)
				}
				for i := 0; i < limit; i++ {
					if t, ok := nextRuns[i].(string); ok {
						fmt.Printf("    %d. %s\n", i+1, t)
					}
				}
			}

			fmt.Println()
			return nil
		},
	}

	return cmd
}
