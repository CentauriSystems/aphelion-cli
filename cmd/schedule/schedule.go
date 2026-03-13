package schedule

import "github.com/spf13/cobra"

func NewScheduleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Manage scheduling for deployed agents",
		Long:  "Set, view, enable, disable, and remove cron schedules for deployed agents",
	}

	cmd.AddCommand(newSetCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newEnableCmd())
	cmd.AddCommand(newDisableCmd())
	cmd.AddCommand(newRemoveCmd())

	return cmd
}
