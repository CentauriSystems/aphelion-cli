package analytics

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
)

func newToolsCmd() *cobra.Command {
	var timeframe string
	var userOnly bool
	var agent string
	var last string

	cmd := &cobra.Command{
		Use:   "tools",
		Short: "Show tool usage analytics",
		Long:  "Display analytics for tool usage and popularity",
		Example: `  # Show global tool usage analytics
  aphelion analytics tools

  # Show only your tool usage
  aphelion analytics tools --user-only

  # Show tool analytics for the past week
  aphelion analytics tools --timeframe week --user-only`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				return fmt.Errorf("authentication required. Please run 'aphelion auth login' first")
			}

			client := api.NewClient()
			
			params := map[string]string{
				"timeframe": timeframe,
				"user_only": fmt.Sprintf("%t", userOnly),
			}
			if agent != "" {
				params["agent"] = agent
			}
			if last != "" {
				params["period"] = last
			}

			var analytics api.Analytics
			if err := client.GetWithQuery("/analytics/tools", params, &analytics); err != nil {
				return fmt.Errorf("failed to get tool analytics: %w", err)
			}

			data := map[string]interface{}{
				"Tool Metrics": map[string]interface{}{
					"Total Executions": analytics.ToolMetrics.TotalExecutions,
					"Unique Tools":     analytics.ToolMetrics.UniqueTools,
				},
			}

			if len(analytics.ToolMetrics.PopularTools) > 0 {
				var popularTools []map[string]interface{}
				for _, tool := range analytics.ToolMetrics.PopularTools {
					popularTools = append(popularTools, map[string]interface{}{
						"Tool":         tool.Tool,
						"Count":        tool.Count,
						"Success Rate": fmt.Sprintf("%.2f%%", tool.SuccessRate*100),
					})
				}
				data["Popular Tools"] = popularTools
			}

			scope := "Global"
			if userOnly {
				scope = "Your"
			}
			utils.PrintInfo("%s tool usage analytics (%s)", scope, timeframe)

			return utils.PrintOutput(data, config.GetOutputFormat())
		},
	}

	cmd.Flags().StringVarP(&timeframe, "timeframe", "t", "day", "time period (hour, day, week, month)")
	cmd.Flags().BoolVar(&userOnly, "user-only", false, "show only current user's tool usage")
	cmd.Flags().StringVar(&agent, "agent", "", "filter by agent name or ID")
	cmd.Flags().StringVar(&last, "last", "", "time period filter (e.g. 7d, 30d)")

	return cmd
}