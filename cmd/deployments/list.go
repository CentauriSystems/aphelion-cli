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

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all deployed agents",
		Long:  "List all agents deployed to the Aphelion cloud with their status and endpoints.",
		Example: `  aphelion deployments list
  aphelion deployments list -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			s := utils.NewSpinner("Fetching deployments...")
			s.Start()

			client := api.NewClient()
			var resp api.DeploymentsResponse
			err := client.Get("/v2/deployments", &resp)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to list deployments: %v", err)
				return err
			}

			if len(resp.Deployments) == 0 {
				fmt.Println("No deployments found.")
				fmt.Println("Deploy an agent with: aphelion deploy")
				return nil
			}

			outputFormat := config.GetOutputFormat()
			if f, _ := cmd.Flags().GetString("output"); f != "" {
				outputFormat = f
			}

			if outputFormat == "json" || outputFormat == "yaml" {
				return utils.PrintOutput(resp.Deployments, outputFormat)
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Agent", "Status", "Endpoint", "Region", "Last Deploy", "Executions (24h)"})
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

			for _, d := range resp.Deployments {
				table.Append([]string{
					d.AgentName,
					d.Status,
					d.Endpoint,
					d.Region,
					d.LastDeployed,
					fmt.Sprintf("%d", d.Executions24),
				})
			}

			table.Render()
			fmt.Printf("\nTotal: %d deployments\n", resp.Total)

			return nil
		},
	}

	return cmd
}
