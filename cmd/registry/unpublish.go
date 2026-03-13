package registry

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newUnpublishCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unpublish <service-id>",
		Short: "Unpublish a service from the marketplace",
		Long:  "Remove a service from the marketplace. The service will no longer be discoverable by other users.",
		Args:  cobra.ExactArgs(1),
		Example: `  # Unpublish a service
  aphelion registry unpublish service-123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				return fmt.Errorf("authentication required. Run: aphelion auth login")
			}

			serviceID := args[0]
			client := api.NewClient()

			spinner := utils.NewSpinner("Unpublishing service...")
			spinner.Start()

			endpoint := fmt.Sprintf("/v2/services/%s/unpublish", serviceID)
			err := client.Post(endpoint, nil, nil)
			spinner.Stop()

			if err != nil {
				return fmt.Errorf("failed to unpublish service: %w", err)
			}

			utils.PrintSuccess("Service %s unpublished successfully", serviceID)
			return nil
		},
	}

	return cmd
}
