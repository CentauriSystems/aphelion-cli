package agents

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newUpdateCmd() *cobra.Command {
	var description string

	cmd := &cobra.Command{
		Use:   "update <name-or-id>",
		Short: "Update an agent identity",
		Long:  "Update the description of an agent identity by name or ID.",
		Example: `  aphelion agents update review-agent --description "Updated description"
  aphelion agents update agt_KFnG9Ad2r9LsCrJdLu9k --description "New desc"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			nameOrID := args[0]

			s := utils.NewSpinner("Updating agent...")
			s.Start()

			client := api.NewClient()
			body := map[string]string{
				"description": description,
			}

			var agent api.AgentIdentity
			err := client.Patch(fmt.Sprintf("/v2/agents/%s", nameOrID), body, &agent)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to update agent %q: %v", nameOrID, err)
				return err
			}

			utils.PrintSuccess("Agent %q updated successfully.", agent.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "New agent description (required)")
	cmd.MarkFlagRequired("description")

	return cmd
}
