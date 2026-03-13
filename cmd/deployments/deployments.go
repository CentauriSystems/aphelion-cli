package deployments

import "github.com/spf13/cobra"

func NewDeploymentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deployments",
		Short: "Manage deployed agents",
		Long:  "List, monitor, and manage agents deployed to the Aphelion cloud.",
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newLogsCmd())
	cmd.AddCommand(newHistoryCmd())
	cmd.AddCommand(newRollbackCmd())
	cmd.AddCommand(newStopCmd())
	cmd.AddCommand(newRedeployCmd())
	cmd.AddCommand(newDeleteCmd())

	return cmd
}
