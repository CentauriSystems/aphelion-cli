package analytics

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

type ExecutionEntry struct {
	ID           string `json:"id"`
	Agent        string `json:"agent"`
	Status       string `json:"status"`
	StartedAt    string `json:"started_at"`
	Duration     string `json:"duration"`
	InputSummary string `json:"input_summary"`
}

type ExecutionsResponse struct {
	Executions []ExecutionEntry `json:"executions"`
}

func newExecutionsCmd() *cobra.Command {
	var agent string
	var status string
	var last string

	cmd := &cobra.Command{
		Use:   "executions",
		Short: "Show execution analytics",
		Long:  "Display analytics for agent executions with optional filtering",
		Example: `  # Show all executions
  aphelion analytics executions

  # Show executions for a specific agent
  aphelion analytics executions --agent review-management-agent

  # Show only failed executions
  aphelion analytics executions --status failed

  # Show executions from the last 7 days
  aphelion analytics executions --last 7d

  # Combine filters
  aphelion analytics executions --agent my-agent --status failed --last 30d`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				return fmt.Errorf("authentication required. Please run 'aphelion auth login' first")
			}

			client := api.NewClient()

			params := map[string]string{}
			if agent != "" {
				params["agent"] = agent
			}
			if status != "" {
				params["status"] = status
			}
			if last != "" {
				params["period"] = last
			}

			var resp ExecutionsResponse
			if err := client.GetWithQuery("/v2/analytics/executions", params, &resp); err != nil {
				return fmt.Errorf("failed to get execution analytics: %w", err)
			}

			if len(resp.Executions) == 0 {
				fmt.Println("No executions found matching the given filters.")
				return nil
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Agent", "Status", "Started", "Duration", "Input Summary"})
			table.SetBorder(false)
			table.SetAutoWrapText(false)
			table.SetColumnSeparator(" ")

			for _, e := range resp.Executions {
				table.Append([]string{
					e.ID,
					e.Agent,
					e.Status,
					e.StartedAt,
					e.Duration,
					e.InputSummary,
				})
			}

			table.Render()
			return nil
		},
	}

	cmd.Flags().StringVar(&agent, "agent", "", "filter by agent name or ID")
	cmd.Flags().StringVar(&status, "status", "", "filter by execution status (e.g. failed, success)")
	cmd.Flags().StringVar(&last, "last", "", "time period filter (e.g. 7d, 30d, 24h)")

	return cmd
}
