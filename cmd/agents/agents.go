package agents

import "github.com/spf13/cobra"

func NewAgentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Manage agent identities",
		Long:  "Create, list, and manage agent identities on the Aphelion platform",
	}

	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newRotateSecretCmd())
	cmd.AddCommand(newSuspendCmd())
	cmd.AddCommand(newActivateCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newInspectCmd())
	AddPermissionCmds(cmd)

	return cmd
}
