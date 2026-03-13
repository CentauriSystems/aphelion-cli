package analytics

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
)

func newSessionsCmd() *cobra.Command {
	var timeframe string
	var userOnly bool
	var agent string
	var last string

	cmd := &cobra.Command{
		Use:   "sessions",
		Short: "Show session analytics",
		Long:  "Display analytics for session usage and activity",
		Example: `  # Show global session analytics
  aphelion analytics sessions

  # Show only your session analytics
  aphelion analytics sessions --user-only

  # Show session analytics for the past month
  aphelion analytics sessions --timeframe month --user-only`,
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
			if err := client.GetWithQuery("/analytics/sessions", params, &analytics); err != nil {
				return fmt.Errorf("failed to get session analytics: %w", err)
			}

			data := map[string]interface{}{
				"Session Metrics": map[string]interface{}{
					"Total Sessions":     analytics.SessionMetrics.TotalSessions,
					"Active Sessions":    analytics.SessionMetrics.ActiveSessions,
					"Average Activities": fmt.Sprintf("%.2f", analytics.SessionMetrics.AverageActivities),
					"Average Duration":   fmt.Sprintf("%.2f minutes", analytics.SessionMetrics.AverageDuration),
				},
			}

			scope := "Global"
			if userOnly {
				scope = "Your"
			}
			utils.PrintInfo("%s session analytics (%s)", scope, timeframe)

			return utils.PrintOutput(data, config.GetOutputFormat())
		},
	}

	cmd.Flags().StringVarP(&timeframe, "timeframe", "t", "day", "time period (hour, day, week, month)")
	cmd.Flags().BoolVar(&userOnly, "user-only", true, "show only current user's sessions")
	cmd.Flags().StringVar(&agent, "agent", "", "filter by agent name or ID")
	cmd.Flags().StringVar(&last, "last", "", "time period filter (e.g. 7d, 30d)")

	return cmd
}