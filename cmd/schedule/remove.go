package schedule

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

func newRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <agent-name>",
		Short: "Remove the schedule for a deployed agent",
		Long:  "Remove the cron schedule entirely from a deployed agent. The agent will no longer run on a schedule.",
		Example: `  aphelion schedule remove review-management-agent`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			agentName := args[0]

			fmt.Printf("Remove schedule for %s? [y/N] ", agentName)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))

			if answer != "y" && answer != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}

			s := utils.NewSpinner("Removing schedule...")
			s.Start()

			client := api.NewClient()
			err := client.Delete(fmt.Sprintf("/v2/agents/%s/schedule", agentName))
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to remove schedule: %v", err)
				return err
			}

			utils.PrintSuccess("Schedule removed for %s", agentName)
			return nil
		},
	}

	return cmd
}
