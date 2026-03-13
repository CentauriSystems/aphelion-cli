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

type grantsResponse struct {
	Permissions []api.AgentPermission `json:"permissions"`
}

func newGrantsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grants <name-or-id>",
		Short: "Show what an agent has been granted access to",
		Long:  "List all agents whose memory the specified agent has been granted access to.",
		Example: `  aphelion agents grants researcher-agent
  aphelion agents grants agt_KFnG9Ad2r9LsCrJdLu9k`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			nameOrID := args[0]

			s := utils.NewSpinner("Fetching grants...")
			s.Start()

			client := api.NewClient()
			var resp grantsResponse
			err := client.Get(fmt.Sprintf("/v2/agents/%s/grants", nameOrID), &resp)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to fetch grants for %q: %v", nameOrID, err)
				return err
			}

			if len(resp.Permissions) == 0 {
				fmt.Printf("%s has not been granted access to any other agent's memory.\n", nameOrID)
				fmt.Println("Grant access with: aphelion agents grant --from", nameOrID, "--to <agent> --actions read")
				return nil
			}

			fmt.Printf("Grants for %s:\n\n", nameOrID)

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Resource Agent", "Actions", "Expires"})
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
					p.ResourceAgent,
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
