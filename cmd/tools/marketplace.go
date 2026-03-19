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
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Category    string      `json:"category"`
	Pricing     interface{} `json:"pricing"`
	Rating      float64     `json:"rating"`
	ToolCount   int         `json:"tool_count"`
	SpecTitle   string      `json:"spec_title"`
}

type marketplaceResponse struct {
	Tools    []marketplaceTool `json:"tools"`
	Services []marketplaceTool `json:"services"`
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
			err := client.GetWithQuery("/services/summary", params, &resp)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to load marketplace: %v", err)
				return err
			}

			// Backend returns "services" key, merge into tools
			if len(resp.Services) > 0 && len(resp.Tools) == 0 {
				resp.Tools = resp.Services
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
			table.SetHeader([]string{"Name", "Description", "Category", "Pricing", "Tools"})
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
				name := t.Name
				if t.SpecTitle != "" {
					name = t.SpecTitle
				}
				desc := t.Description
				if len(desc) > 50 {
					desc = desc[:47] + "..."
				}
				pricing := "free"
				if pm, ok := t.Pricing.(map[string]interface{}); ok {
					if free, ok := pm["free"].(bool); ok && !free {
						pricing = "paid"
					}
				} else if ps, ok := t.Pricing.(string); ok {
					pricing = ps
				}
				tools := fmt.Sprintf("%d", t.ToolCount)
				table.Append([]string{name, desc, t.Category, pricing, tools})
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
