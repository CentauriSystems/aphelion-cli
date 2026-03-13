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

type subscribedTool struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type subscribedToolsResponse struct {
	Tools []subscribedTool `json:"tools"`
}

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tools the current agent is subscribed to",
		Long:  "Show all tools that the current project's agent is subscribed to.",
		Example: `  aphelion tools list
  aphelion tools list -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsProjectDir() {
				utils.PrintError("Not in an Aphelion project directory.")
				fmt.Println("Run this command from a directory with .aphelion/config.yaml")
				fmt.Println("Or create a new project: aphelion agent init")
				return fmt.Errorf("not in project directory")
			}

			agentID := config.GetAgentID()
			if agentID == "" {
				utils.PrintError("No agent ID found in project config.")
				fmt.Println("Create an agent identity: aphelion agents create --name <name>")
				return fmt.Errorf("no agent ID")
			}

			s := utils.NewSpinner("Fetching subscribed tools...")
			s.Start()

			client := api.NewClient()
			endpoint := fmt.Sprintf("/v2/agents/%s/tools", agentID)
			var resp subscribedToolsResponse
			err := client.Get(endpoint, &resp)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to list tools: %v", err)
				return err
			}

			if len(resp.Tools) == 0 {
				fmt.Println("No tools subscribed.")
				fmt.Println("Search for tools: aphelion tools search <query>")
				fmt.Println("Browse marketplace: aphelion tools marketplace")
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
			table.SetHeader([]string{"Name", "Status"})
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
				table.Append([]string{t.Name, t.Status})
			}

			table.Render()
			fmt.Printf("\nTotal: %d tools\n", len(resp.Tools))

			return nil
		},
	}

	return cmd
}
