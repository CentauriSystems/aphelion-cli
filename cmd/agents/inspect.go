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

func newInspectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect <name-or-id>",
		Short: "Inspect an agent in detail",
		Long:  "Show detailed information about an agent: tool subscriptions, memory, permissions, deployment, and recent executions.",
		Example: `  aphelion agents inspect review-agent
  aphelion agents inspect agt_KFnG9Ad2r9LsCrJdLu9k`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			nameOrID := args[0]

			s := utils.NewSpinner("Inspecting agent...")
			s.Start()

			client := api.NewClient()
			var inspection api.AgentInspection
			err := client.Get(fmt.Sprintf("/v2/agents/%s/inspect", nameOrID), &inspection)
			s.Stop()

			if err != nil {
				utils.PrintError("Agent %q not found.\nList your agents: aphelion agents list", nameOrID)
				return err
			}

			bold := color.New(color.Bold)
			dim := color.New(color.Faint)

			fmt.Println()
			bold.Println(inspection.Agent.Name)
			fmt.Println(strings.Repeat("=", 50))

			// Agent info
			fmt.Printf("  ID:          %s\n", inspection.Agent.ID)
			fmt.Printf("  Description: %s\n", inspection.Agent.Description)
			fmt.Printf("  Status:      %s\n", inspection.Agent.Status)
			fmt.Printf("  Created:     %s\n", inspection.Agent.CreatedAt.Format("2006-01-02 15:04:05"))
			if !inspection.Agent.LastActive.IsZero() {
				fmt.Printf("  Last Active: %s\n", inspection.Agent.LastActive.Format("2006-01-02 15:04:05"))
			}
			fmt.Println()

			// Tool subscriptions
			bold.Println("Tool Subscriptions")
			fmt.Println(strings.Repeat("-", 50))
			if len(inspection.Tools) == 0 {
				dim.Println("  No tools subscribed")
			} else {
				for _, tool := range inspection.Tools {
					fmt.Printf("  - %s\n", tool)
				}
			}
			fmt.Println()

			// Memory
			bold.Println("Memory")
			fmt.Println(strings.Repeat("-", 50))
			fmt.Printf("  Entries: %d\n", inspection.MemoryCount)
			fmt.Println()

			// Permissions
			bold.Println("Permissions")
			fmt.Println(strings.Repeat("-", 50))
			if len(inspection.Permissions) == 0 {
				dim.Println("  No permissions granted")
			} else {
				for _, perm := range inspection.Permissions {
					actions := strings.Join(perm.Actions, ", ")
					fmt.Printf("  %s -> %s  [%s]", perm.GranteeAgent, perm.ResourceAgent, actions)
					if perm.ExpiresAt != "" {
						fmt.Printf("  expires: %s", perm.ExpiresAt)
					}
					fmt.Println()
				}
			}
			fmt.Println()

			// Deployment
			bold.Println("Deployment")
			fmt.Println(strings.Repeat("-", 50))
			if inspection.Deployment == nil {
				dim.Println("  Not deployed")
			} else {
				fmt.Printf("  Status:       %s\n", inspection.Deployment.Status)
				fmt.Printf("  Endpoint:     %s\n", inspection.Deployment.Endpoint)
				fmt.Printf("  Region:       %s\n", inspection.Deployment.Region)
				fmt.Printf("  Last Deploy:  %s\n", inspection.Deployment.LastDeployed)
			}
			fmt.Println()

			// Recent executions
			bold.Println("Recent Executions")
			fmt.Println(strings.Repeat("-", 50))
			if len(inspection.RecentExecs) == 0 {
				dim.Println("  No recent executions")
			} else {
				for _, exec := range inspection.RecentExecs {
					fmt.Printf("  %s  %-8s  %s  %s\n", exec.ID, exec.Status, exec.StartedAt, exec.Duration)
				}
			}
			fmt.Println()

			return nil
		},
	}

	return cmd
}
