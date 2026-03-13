package agents

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <name-or-id>",
		Short: "Get details of an agent identity",
		Long:  "Get full details of an agent identity by name or ID.",
		Example: `  aphelion agents get review-agent
  aphelion agents get agt_KFnG9Ad2r9LsCrJdLu9k
  aphelion agents get review-agent -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			nameOrID := args[0]

			s := utils.NewSpinner("Fetching agent details...")
			s.Start()

			client := api.NewClient()
			var agent api.AgentIdentity
			err := client.Get(fmt.Sprintf("/v2/agents/%s", nameOrID), &agent)
			s.Stop()

			if err != nil {
				utils.PrintError("Agent %q not found.\nList your agents: aphelion agents list", nameOrID)
				return err
			}

			outputFormat := config.GetOutputFormat()
			if f, _ := cmd.Flags().GetString("output"); f != "" {
				outputFormat = f
			}

			if outputFormat == "json" || outputFormat == "yaml" {
				return utils.PrintOutput(agent, outputFormat)
			}

			bold := color.New(color.Bold)
			fmt.Println()
			bold.Println(agent.Name)
			fmt.Println(strings.Repeat("-", 50))
			fmt.Printf("  ID:           %s\n", agent.ID)
			fmt.Printf("  Description:  %s\n", agent.Description)
			fmt.Printf("  Status:       %s\n", agent.Status)

			if agent.ClientID != "" {
				// Mask client ID — show prefix only
				masked := agent.ClientID
				if len(masked) > 8 {
					masked = masked[:8] + "..."
				}
				fmt.Printf("  Client ID:    %s\n", masked)
			}

			fmt.Printf("  Created:      %s\n", agent.CreatedAt.Format("2006-01-02 15:04:05"))

			if !agent.LastActive.IsZero() {
				fmt.Printf("  Last Active:  %s\n", agent.LastActive.Format("2006-01-02 15:04:05"))
			} else {
				fmt.Printf("  Last Active:  -\n")
			}

			fmt.Println(strings.Repeat("-", 50))
			fmt.Println()

			return nil
		},
	}

	return cmd
}
