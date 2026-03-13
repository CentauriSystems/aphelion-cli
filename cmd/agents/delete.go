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

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name-or-id>",
		Short: "Delete an agent identity",
		Long:  "Permanently delete an agent identity. This cannot be undone.",
		Example: `  aphelion agents delete review-agent
  aphelion agents delete agt_KFnG9Ad2r9LsCrJdLu9k`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			nameOrID := args[0]

			fmt.Printf("Type the agent name to confirm deletion: ")
			reader := bufio.NewReader(os.Stdin)
			confirmation, _ := reader.ReadString('\n')
			confirmation = strings.TrimSpace(confirmation)

			if confirmation != nameOrID {
				utils.PrintError("Confirmation does not match. Deletion aborted.")
				return nil
			}

			s := utils.NewSpinner("Deleting agent...")
			s.Start()

			client := api.NewClient()
			err := client.Delete(fmt.Sprintf("/v2/agents/%s", nameOrID))
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to delete agent %q: %v", nameOrID, err)
				return err
			}

			utils.PrintSuccess("Agent %q deleted.", nameOrID)

			return nil
		},
	}

	return cmd
}
