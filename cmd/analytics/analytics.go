package analytics

import (
	"github.com/spf13/cobra"
)

func NewAnalyticsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analytics",
		Short: "Analytics and usage metrics",
		Long:  "View usage analytics and metrics for your account",
	}

	cmd.AddCommand(newUserCmd())
	cmd.AddCommand(newToolsCmd())
	cmd.AddCommand(newSessionsCmd())
	cmd.AddCommand(newExecutionsCmd())
	cmd.AddCommand(newEarningsCmd())

	return cmd
}