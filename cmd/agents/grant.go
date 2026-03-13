package agents

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newGrantCmd() *cobra.Command {
	var fromAgent string
	var toAgent string
	var actions string
	var expires string

	cmd := &cobra.Command{
		Use:   "grant",
		Short: "Grant an agent permission to access another agent's memory",
		Long:  "Grant read or read/write access from one agent to another agent's memory. Enables multi-agent workflows.",
		Example: `  aphelion agents grant --from researcher-agent --to writer-agent --actions read
  aphelion agents grant --from researcher-agent --to writer-agent --actions read,write --expires 2026-12-31`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			actionList := strings.Split(actions, ",")
			for i, a := range actionList {
				actionList[i] = strings.TrimSpace(a)
			}

			s := utils.NewSpinner("Granting permissions...")
			s.Start()

			client := api.NewClient()
			req := api.GrantRequest{
				FromAgent: fromAgent,
				ToAgent:   toAgent,
				Actions:   actionList,
				ExpiresAt: expires,
			}

			err := client.Post("/v2/agents/permissions/grant", req, nil)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to grant permissions: %v", err)
				return err
			}

			utils.PrintSuccess("Granted [%s] access from %s to %s's memory", strings.Join(actionList, ", "), fromAgent, toAgent)

			if expires != "" {
				fmt.Printf("  Expires: %s\n", expires)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&fromAgent, "from", "", "Grantee agent name or ID (required)")
	cmd.Flags().StringVar(&toAgent, "to", "", "Resource agent name or ID (required)")
	cmd.Flags().StringVar(&actions, "actions", "", "Comma-separated actions: read, write (required)")
	cmd.Flags().StringVar(&expires, "expires", "", "Expiration date (e.g. 2026-12-31)")
	cmd.MarkFlagRequired("from")
	cmd.MarkFlagRequired("to")
	cmd.MarkFlagRequired("actions")

	return cmd
}
