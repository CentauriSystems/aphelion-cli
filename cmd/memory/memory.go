package memory

import (
	"github.com/spf13/cobra"
)

func NewMemoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Memory management commands",
		Long:  "Manage and search through your AI agent memories",
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newSearchCmd())
	cmd.AddCommand(newStatsCmd())
	cmd.AddCommand(newClearCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newSetCmd())
	cmd.AddCommand(newDeleteCmd())

	return cmd
}