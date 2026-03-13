package analytics

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

type EarningsBreakdown struct {
	Source string  `json:"source"`
	Amount float64 `json:"amount"`
}

type EarningsResponse struct {
	Total     float64             `json:"total"`
	Period    string              `json:"period"`
	Breakdown []EarningsBreakdown `json:"breakdown"`
}

func newEarningsCmd() *cobra.Command {
	var last string

	cmd := &cobra.Command{
		Use:   "earnings",
		Short: "Show earnings analytics",
		Long:  "Display earnings from published APIs and agents",
		Example: `  # Show earnings overview
  aphelion analytics earnings

  # Show earnings for the last 30 days
  aphelion analytics earnings --last 30d`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				return fmt.Errorf("authentication required. Please run 'aphelion auth login' first")
			}

			client := api.NewClient()

			params := map[string]string{}
			if last != "" {
				params["period"] = last
			}

			var resp EarningsResponse
			if err := client.GetWithQuery("/v2/analytics/earnings", params, &resp); err != nil {
				return fmt.Errorf("failed to get earnings analytics: %w", err)
			}

			period := resp.Period
			if period == "" {
				period = "all time"
			}

			fmt.Printf("\nEarnings (%s)\n", period)
			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			fmt.Printf("  Total: $%.2f\n\n", resp.Total)

			if len(resp.Breakdown) > 0 {
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Source", "Amount"})
				table.SetBorder(false)
				table.SetColumnSeparator(" ")

				for _, b := range resp.Breakdown {
					table.Append([]string{
						b.Source,
						fmt.Sprintf("$%.2f", b.Amount),
					})
				}

				table.Render()
			} else {
				fmt.Println("  No earnings breakdown available.")
			}

			fmt.Println()
			return nil
		},
	}

	cmd.Flags().StringVar(&last, "last", "", "time period filter (e.g. 7d, 30d)")

	return cmd
}
