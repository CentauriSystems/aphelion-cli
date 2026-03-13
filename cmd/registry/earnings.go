package registry

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

type earningsResponse struct {
	Total    float64          `json:"total"`
	Currency string           `json:"currency"`
	Services []serviceEarning `json:"services"`
}

type serviceEarning struct {
	Name     string  `json:"name"`
	Earnings float64 `json:"earnings"`
	Calls    int     `json:"calls"`
}

func newEarningsCmd() *cobra.Command {
	var service string
	var last string

	cmd := &cobra.Command{
		Use:   "earnings",
		Short: "View earnings for published APIs",
		Long:  "View earnings from API calls to your published services.",
		Example: `  # View all earnings
  aphelion registry earnings

  # View earnings for a specific service
  aphelion registry earnings --service service-123

  # View earnings for last 30 days
  aphelion registry earnings --last 30d`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				return fmt.Errorf("authentication required. Run: aphelion auth login")
			}

			client := api.NewClient()

			spinner := utils.NewSpinner("Fetching earnings...")
			spinner.Start()

			params := map[string]string{}
			if service != "" {
				params["service"] = service
			}
			if last != "" {
				params["last"] = last
			}

			var resp earningsResponse
			err := client.GetWithQuery("/v2/services/earnings", params, &resp)
			spinner.Stop()

			if err != nil {
				return fmt.Errorf("failed to fetch earnings: %w", err)
			}

			outputFormat := config.GetOutputFormat()
			if f, _ := cmd.Flags().GetString("output"); f != "" {
				outputFormat = f
			}

			if outputFormat == "json" || outputFormat == "yaml" {
				return utils.PrintOutput(resp, outputFormat)
			}

			if len(resp.Services) == 0 {
				utils.PrintInfo("No earnings data found")
				return nil
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Service", "Earnings", "Calls"})
			table.SetAutoWrapText(false)
			table.SetAutoFormatHeaders(true)
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetCenterSeparator("")
			table.SetColumnSeparator("")
			table.SetRowSeparator("")
			table.SetHeaderLine(false)
			table.SetBorder(false)
			table.SetTablePadding("  ")

			for _, s := range resp.Services {
				table.Append([]string{
					s.Name,
					fmt.Sprintf("%.2f %s", s.Earnings, resp.Currency),
					fmt.Sprintf("%d", s.Calls),
				})
			}

			table.Render()
			fmt.Printf("\nTotal: %.2f %s\n", resp.Total, resp.Currency)

			return nil
		},
	}

	cmd.Flags().StringVar(&service, "service", "", "filter by service ID")
	cmd.Flags().StringVar(&last, "last", "", "time period (e.g. 7d, 30d)")

	return cmd
}
