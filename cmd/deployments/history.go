package deployments

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newHistoryCmd() *cobra.Command {
	var limit int
	var status string

	cmd := &cobra.Command{
		Use:   "history <agent-name>",
		Short: "View execution history for a deployed agent",
		Long:  "View recent execution history for a deployed agent including input/output summaries and duration.",
		Args:  cobra.ExactArgs(1),
		Example: `  aphelion deployments history my-agent
  aphelion deployments history my-agent --limit 50
  aphelion deployments history my-agent --status failed`,
		RunE: func(cmd *cobra.Command, args []string) error {
			agentName := args[0]

			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			s := utils.NewSpinner("Fetching execution history...")
			s.Start()

			client := api.NewClient()
			endpoint := fmt.Sprintf("/v2/agents/%s/executions", agentName)
			params := map[string]string{
				"limit": fmt.Sprintf("%d", limit),
			}
			if status != "" {
				params["status"] = status
			}

			var resp api.ExecutionsResponse
			err := client.GetWithQuery(endpoint, params, &resp)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to fetch execution history: %v", err)
				return err
			}

			if len(resp.Executions) == 0 {
				fmt.Println("No executions found.")
				return nil
			}

			outputFormat := config.GetOutputFormat()
			if f, _ := cmd.Flags().GetString("output"); f != "" {
				outputFormat = f
			}

			if outputFormat == "json" || outputFormat == "yaml" {
				return utils.PrintOutput(resp.Executions, outputFormat)
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Timestamp", "Input Summary", "Output Summary", "Duration", "Status"})
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

			for _, e := range resp.Executions {
				table.Append([]string{
					e.Timestamp,
					e.InputSummary,
					e.OutputSummary,
					e.Duration,
					e.Status,
				})
			}

			table.Render()
			fmt.Printf("\nShowing %d of %d executions\n", len(resp.Executions), resp.Total)

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "Number of executions to retrieve")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status (e.g. success, failed)")

	return cmd
}
