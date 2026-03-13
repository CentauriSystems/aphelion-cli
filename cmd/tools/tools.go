package tools

import (
	"github.com/spf13/cobra"
)

func NewToolsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tools",
		Short: "Tool discovery and execution commands",
		Long:  "Discover, describe, and test tools available in Aphelion Gateway",
	}

	cmd.AddCommand(newDescribeCmd())
	cmd.AddCommand(newTryCmd())
	cmd.AddCommand(newSearchCmd())
	cmd.AddCommand(newSubscribeCmd())
	cmd.AddCommand(newUnsubscribeCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newMarketplaceCmd())

	return cmd
}