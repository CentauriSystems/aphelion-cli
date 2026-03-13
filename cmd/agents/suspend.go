package agents

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newSuspendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "suspend <name-or-id>",
		Short: "Suspend an agent",
		Long:  "Suspend an agent identity. All executions will return 503 while suspended.",
		Example: `  aphelion agents suspend review-agent
  aphelion agents suspend agt_KFnG9Ad2r9LsCrJdLu9k`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			nameOrID := args[0]

			s := utils.NewSpinner("Suspending agent...")
			s.Start()

			client := api.NewClient()
			err := client.Post(fmt.Sprintf("/v2/agents/%s/suspend", nameOrID), nil, nil)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to suspend agent %q: %v", nameOrID, err)
				return err
			}

			utils.PrintSuccess("Agent %q suspended. All executions will return 503.", nameOrID)

			return nil
		},
	}

	return cmd
}
