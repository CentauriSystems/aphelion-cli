package deployments

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

func newRollbackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback <agent-name>",
		Short: "Rollback to the previous deployment version",
		Long:  "Rollback a deployed agent to the previous deployment version.",
		Args:  cobra.ExactArgs(1),
		Example: `  aphelion deployments rollback my-agent`,
		RunE: func(cmd *cobra.Command, args []string) error {
			agentName := args[0]

			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			fmt.Printf("This will rollback agent \"%s\" to the previous deployment version.\n", agentName)
			fmt.Print("Continue? [y/N]: ")

			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))

			if answer != "y" && answer != "yes" {
				fmt.Println("Rollback cancelled.")
				return nil
			}

			s := utils.NewSpinner("Rolling back deployment...")
			s.Start()

			client := api.NewClient()
			endpoint := fmt.Sprintf("/v2/agents/%s/rollback", agentName)
			var resp api.RollbackResponse
			err := client.Post(endpoint, nil, &resp)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to rollback: %v", err)
				return err
			}

			utils.PrintSuccess("Rollback complete.")
			fmt.Printf("  Version: %s\n", resp.Version)
			fmt.Printf("  Status:  %s\n", resp.Status)
			if resp.Message != "" {
				fmt.Printf("  Message: %s\n", resp.Message)
			}

			return nil
		},
	}

	return cmd
}
