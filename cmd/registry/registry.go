package registry

import (
	"github.com/spf13/cobra"
)

func NewRegistryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "API service registry management",
		Long:  "Manage API service registration and discovery",
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newMyServicesCmd())
	cmd.AddCommand(newAddOpenAPICmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newPublishCmd())
	cmd.AddCommand(newUnpublishCmd())
	cmd.AddCommand(newEarningsCmd())

	return cmd
}