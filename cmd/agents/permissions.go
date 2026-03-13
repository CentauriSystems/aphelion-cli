package agents

import (
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

type permissionsResponse struct {
	Permissions []api.AgentPermission `json:"permissions"`
}

func newPermissionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "permissions <name-or-id>",
		Short: "Show who has been granted access to an agent's memory",
		Long:  "List all agents that have been granted permission to access the specified agent's memory.",
		Example: `  aphelion agents permissions review-agent
  aphelion agents permissions agt_KFnG9Ad2r9LsCrJdLu9k`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			nameOrID := args[0]

			s := utils.NewSpinner("Fetching permissions...")
			s.Start()

			client := api.NewClient()
			var resp permissionsResponse
			err := client.Get(fmt.Sprintf("/v2/agents/%s/permissions", nameOrID), &resp)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to fetch permissions for %q: %v", nameOrID, err)
				return err
			}

			if len(resp.Permissions) == 0 {
				fmt.Printf("No agents have been granted access to %s's memory.\n", nameOrID)
				fmt.Println("Grant access with: aphelion agents grant --from <agent> --to", nameOrID, "--actions read")
				return nil
			}

			fmt.Printf("Permissions for %s:\n\n", nameOrID)

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Grantee", "Actions", "Expires"})
			table.SetAutoWrapText(false)
			table.SetAutoFormatHeaders(true)
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetCenterSeparator("")
			table.SetColumnSeparator("")
			table.SetRowSeparator("")
			table.SetHeaderLine(false)
			table.SetBorder(false)
			table.SetTablePadding("  ")

			for _, p := range resp.Permissions {
				expires := "-"
				if p.ExpiresAt != "" {
					expires = p.ExpiresAt
				}
				table.Append([]string{
					p.GranteeAgent,
					strings.Join(p.Actions, ", "),
					expires,
				})
			}

			table.Render()

			return nil
		},
	}

	return cmd
}
