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

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <agent-name>",
		Short: "Delete a deployment",
		Long:  "Delete a deployed agent. This stops serving requests and removes the endpoint.",
		Args:  cobra.ExactArgs(1),
		Example: `  aphelion deployments delete my-agent`,
		RunE: func(cmd *cobra.Command, args []string) error {
			agentName := args[0]

			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			fmt.Printf("This will permanently delete the deployment for \"%s\".\n", agentName)
			fmt.Printf("The agent will stop serving requests and its endpoint will be removed.\n\n")
			fmt.Printf("Type the agent name to confirm: ")

			reader := bufio.NewReader(os.Stdin)
			confirmation, _ := reader.ReadString('\n')
			confirmation = strings.TrimSpace(confirmation)

			if confirmation != agentName {
				utils.PrintError("Agent name does not match. Deletion cancelled.")
				return fmt.Errorf("confirmation failed")
			}

			s := utils.NewSpinner("Deleting deployment...")
			s.Start()

			client := api.NewClient()
			endpoint := fmt.Sprintf("/v2/agents/%s/deployment", agentName)
			err := client.Delete(endpoint)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to delete deployment: %v", err)
				return err
			}

			utils.PrintSuccess("Deployment deleted for \"%s\".", agentName)
			fmt.Println("The agent identity still exists. To delete the agent entirely, use: aphelion agents delete " + agentName)

			return nil
		},
	}

	return cmd
}
