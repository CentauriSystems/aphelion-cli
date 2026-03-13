package registry

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newPublishCmd() *cobra.Command {
	var visibility string

	cmd := &cobra.Command{
		Use:   "publish <service-id>",
		Short: "Publish a service to the marketplace",
		Long:  "Publish a service with the specified visibility (public or private).",
		Args:  cobra.ExactArgs(1),
		Example: `  # Publish a service as public
  aphelion registry publish service-123 --visibility public

  # Publish a service as private
  aphelion registry publish service-123 --visibility private`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				return fmt.Errorf("authentication required. Run: aphelion auth login")
			}

			serviceID := args[0]

			if visibility != "public" && visibility != "private" {
				return fmt.Errorf("--visibility must be \"public\" or \"private\"")
			}

			client := api.NewClient()

			spinner := utils.NewSpinner("Publishing service...")
			spinner.Start()

			endpoint := fmt.Sprintf("/v2/services/%s/publish", serviceID)
			body := map[string]string{
				"visibility": visibility,
			}
			err := client.Post(endpoint, body, nil)
			spinner.Stop()

			if err != nil {
				return fmt.Errorf("failed to publish service: %w", err)
			}

			utils.PrintSuccess("Service %s published with visibility: %s", serviceID, visibility)
			return nil
		},
	}

	cmd.Flags().StringVar(&visibility, "visibility", "", "visibility level: public or private")
	_ = cmd.MarkFlagRequired("visibility")

	return cmd
}
