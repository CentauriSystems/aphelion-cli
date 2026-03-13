package agents

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
		Short: "List all agent identities",
		Long:  "List all agent identities on your Aphelion account.",
		Example: `  aphelion agents list
  aphelion agents list -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			s := utils.NewSpinner("Fetching agents...")
			s.Start()

			client := api.NewClient()
			var resp api.AgentsResponse
			err := client.Get("/v2/agents", &resp)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to list agents: %v", err)
				return err
			}

			if len(resp.Agents) == 0 {
				fmt.Println("No agents found.")
				fmt.Println("Create one with: aphelion agents create --name <name> --description <desc>")
				return nil
			}

			outputFormat := config.GetOutputFormat()
			if f, _ := cmd.Flags().GetString("output"); f != "" {
				outputFormat = f
			}

			if outputFormat == "json" || outputFormat == "yaml" {
				return utils.PrintOutput(resp.Agents, outputFormat)
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "ID", "Status", "Created", "Last Active"})
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

			for _, a := range resp.Agents {
				created := a.CreatedAt.Format("2006-01-02")
				lastActive := "-"
				if !a.LastActive.IsZero() {
					lastActive = a.LastActive.Format("2006-01-02 15:04")
				}
				table.Append([]string{a.Name, a.ID, a.Status, created, lastActive})
			}

			table.Render()
			fmt.Printf("\nTotal: %d agents\n", resp.Total)

			return nil
		},
	}

	return cmd
}
