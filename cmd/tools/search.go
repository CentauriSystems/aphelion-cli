package tools

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

type toolSearchResult struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Pricing     string `json:"pricing"`
}

type toolSearchResponse struct {
	Tools []toolSearchResult `json:"tools"`
}

func newSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for tools in the marketplace",
		Long:  "Search for tools by keyword in the Aphelion marketplace.",
		Example: `  aphelion tools search "sms messaging"
  aphelion tools search "email delivery"
  aphelion tools search "payment processing"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			s := utils.NewSpinner("Searching tools...")
			s.Start()

			client := api.NewClient()
			var resp toolSearchResponse
			err := client.GetWithQuery("/search/tools", map[string]string{
				"q": query,
			}, &resp)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to search tools: %v", err)
				return err
			}

			if len(resp.Tools) == 0 {
				fmt.Printf("No tools found for query: %q\n", query)
				fmt.Println("Browse the marketplace: aphelion tools marketplace")
				return nil
			}

			outputFormat := config.GetOutputFormat()
			if f, _ := cmd.Flags().GetString("output"); f != "" {
				outputFormat = f
			}

			if outputFormat == "json" || outputFormat == "yaml" {
				return utils.PrintOutput(resp.Tools, outputFormat)
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Description", "Category", "Pricing"})
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

			for _, t := range resp.Tools {
				desc := t.Description
				if len(desc) > 60 {
					desc = desc[:57] + "..."
				}
				table.Append([]string{t.Name, desc, t.Category, t.Pricing})
			}

			table.Render()
			fmt.Printf("\nFound %d tools matching %q\n", len(resp.Tools), query)

			return nil
		},
	}

	return cmd
}
