package agents

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newActivateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate <name-or-id>",
		Short: "Activate a suspended agent",
		Long:  "Activate a previously suspended agent identity, restoring normal execution.",
		Example: `  aphelion agents activate review-agent
  aphelion agents activate agt_KFnG9Ad2r9LsCrJdLu9k`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			nameOrID := args[0]

			s := utils.NewSpinner("Activating agent...")
			s.Start()

			client := api.NewClient()
			err := client.Post(fmt.Sprintf("/v2/agents/%s/activate", nameOrID), nil, nil)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to activate agent %q: %v", nameOrID, err)
				return err
			}

			utils.PrintSuccess("Agent %q activated.", nameOrID)

			return nil
		},
	}

	return cmd
}
