package agents

import "github.com/spf13/cobra"

// AddPermissionCmds registers the permission-related subcommands on the agents command.
// Call this from NewAgentsCmd or from root command setup.
func AddPermissionCmds(cmd *cobra.Command) {
	cmd.AddCommand(newGrantCmd())
	cmd.AddCommand(newRevokeCmd())
	cmd.AddCommand(newPermissionsCmd())
	cmd.AddCommand(newGrantsCmd())
}
