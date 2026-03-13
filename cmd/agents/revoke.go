package agents

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newRevokeCmd() *cobra.Command {
	var fromAgent string
	var toAgent string

	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke an agent's permission to access another agent's memory",
		Long:  "Revoke previously granted access from one agent to another agent's memory.",
		Example: `  aphelion agents revoke --from researcher-agent --to writer-agent`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			// Prompt for confirmation
			fmt.Printf("Revoke access from %s to %s's memory? [y/N] ", fromAgent, toAgent)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))

			if answer != "y" && answer != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}

			s := utils.NewSpinner("Revoking permissions...")
			s.Start()

			client := api.NewClient()
			req := map[string]string{
				"from_agent": fromAgent,
				"to_agent":   toAgent,
			}

			err := client.Post("/v2/agents/permissions/revoke", req, nil)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to revoke permissions: %v", err)
				return err
			}

			utils.PrintSuccess("Revoked access from %s to %s's memory", fromAgent, toAgent)

			return nil
		},
	}

	cmd.Flags().StringVar(&fromAgent, "from", "", "Grantee agent name or ID (required)")
	cmd.Flags().StringVar(&toAgent, "to", "", "Resource agent name or ID (required)")
	cmd.MarkFlagRequired("from")
	cmd.MarkFlagRequired("to")

	return cmd
}
