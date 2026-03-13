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

type marketplaceTool struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Pricing     string  `json:"pricing"`
	Rating      float64 `json:"rating"`
}

type marketplaceResponse struct {
	Tools []marketplaceTool `json:"tools"`
}

func newMarketplaceCmd() *cobra.Command {
	var (
		category string
		free     bool
		paid     bool
	)

	cmd := &cobra.Command{
		Use:   "marketplace",
		Short: "Browse the tool marketplace",
		Long:  "Browse available tools in the Aphelion marketplace with optional filters.",
		Example: `  aphelion tools marketplace
  aphelion tools marketplace --category communication
  aphelion tools marketplace --free
  aphelion tools marketplace --paid
  aphelion tools marketplace --category data --paid`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := utils.NewSpinner("Loading marketplace...")
			s.Start()

			params := map[string]string{}
			if category != "" {
				params["category"] = category
			}
			if free {
				params["pricing"] = "free"
			}
			if paid {
				params["pricing"] = "paid"
			}

			client := api.NewClient()
			var resp marketplaceResponse
			err := client.GetWithQuery("/v2/tools/marketplace", params, &resp)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to load marketplace: %v", err)
				return err
			}

			if len(resp.Tools) == 0 {
				fmt.Println("No tools found matching your filters.")
				fmt.Println("Try: aphelion tools marketplace")
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
			table.SetHeader([]string{"Name", "Description", "Category", "Pricing", "Rating"})
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
				if len(desc) > 50 {
					desc = desc[:47] + "..."
				}
				rating := "-"
				if t.Rating > 0 {
					rating = fmt.Sprintf("%.1f", t.Rating)
				}
				table.Append([]string{t.Name, desc, t.Category, t.Pricing, rating})
			}

			table.Render()
			fmt.Printf("\nTotal: %d tools\n", len(resp.Tools))
			fmt.Println("Subscribe to a tool: aphelion tools subscribe <name>")

			return nil
		},
	}

	cmd.Flags().StringVar(&category, "category", "", "Filter by category (e.g. communication, data, ai)")
	cmd.Flags().BoolVar(&free, "free", false, "Show only free tools")
	cmd.Flags().BoolVar(&paid, "paid", false, "Show only paid tools")

	return cmd
}
