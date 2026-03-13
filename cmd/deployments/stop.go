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

func newStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop <agent-name>",
		Short: "Stop a deployed agent",
		Long:  "Stop a deployed agent from serving requests. The agent remains registered but its endpoint will return 503.",
		Args:  cobra.ExactArgs(1),
		Example: `  aphelion deployments stop my-agent`,
		RunE: func(cmd *cobra.Command, args []string) error {
			agentName := args[0]

			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			fmt.Printf("This will stop agent \"%s\" from serving requests.\n", agentName)
			fmt.Print("Continue? [y/N]: ")

			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))

			if answer != "y" && answer != "yes" {
				fmt.Println("Stop cancelled.")
				return nil
			}

			s := utils.NewSpinner("Stopping agent...")
			s.Start()

			client := api.NewClient()
			endpoint := fmt.Sprintf("/v2/agents/%s/stop", agentName)
			err := client.Post(endpoint, nil, nil)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to stop agent: %v", err)
				return err
			}

			utils.PrintSuccess("Agent stopped. Endpoint will return 503.")
			fmt.Println("To restart, use: aphelion deployments redeploy " + agentName)

			return nil
		},
	}

	return cmd
}
