package analytics

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
)

func newUserCmd() *cobra.Command {
	var timeframe string
	var last string

	cmd := &cobra.Command{
		Use:   "user",
		Short: "Show user analytics",
		Long:  "Display user-specific analytics including request metrics and usage statistics",
		Example: `  # Show user analytics for the current day
  aphelion analytics user

  # Show user analytics for the past week
  aphelion analytics user --timeframe week

  # Show user analytics in JSON format
  aphelion analytics user --timeframe month --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				return fmt.Errorf("authentication required. Please run 'aphelion auth login' first")
			}

			client := api.NewClient()
			
			params := map[string]string{
				"timeframe": timeframe,
			}
			if last != "" {
				params["period"] = last
			}

			var analytics api.Analytics
			if err := client.GetWithQuery("/analytics/user", params, &analytics); err != nil {
				return fmt.Errorf("failed to get user analytics: %w", err)
			}

			data := map[string]interface{}{
				"Request Metrics": map[string]interface{}{
					"Total Requests":   analytics.RequestMetrics.TotalRequests,
					"Successful Count": analytics.RequestMetrics.SuccessfulCount,
					"Error Count":      analytics.RequestMetrics.ErrorCount,
					"Success Rate":     fmt.Sprintf("%.2f%%", analytics.RequestMetrics.SuccessRate*100),
					"Average Time":     fmt.Sprintf("%.2fms", analytics.RequestMetrics.AverageTime),
				},
				"Session Metrics": map[string]interface{}{
					"Total Sessions":     analytics.SessionMetrics.TotalSessions,
					"Active Sessions":    analytics.SessionMetrics.ActiveSessions,
					"Average Activities": fmt.Sprintf("%.2f", analytics.SessionMetrics.AverageActivities),
					"Average Duration":   fmt.Sprintf("%.2f minutes", analytics.SessionMetrics.AverageDuration),
				},
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

			return utils.PrintOutput(data, config.GetOutputFormat())
		},
	}

	cmd.Flags().StringVarP(&timeframe, "timeframe", "t", "day", "time period (hour, day, week, month)")
	cmd.Flags().StringVar(&last, "last", "", "time period filter (e.g. 7d, 30d)")

	return cmd
}