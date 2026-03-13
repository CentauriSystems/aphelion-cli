package registry

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newUpdateCmd() *cobra.Command {
	var price float64
	var description string

	cmd := &cobra.Command{
		Use:   "update <service-id>",
		Short: "Update a service",
		Long:  "Update the price or description of a service you own.",
		Args:  cobra.ExactArgs(1),
		Example: `  # Update service price
  aphelion registry update service-123 --price 0.002

  # Update service description
  aphelion registry update service-123 --description "Updated description"

  # Update both
  aphelion registry update service-123 --price 0.005 --description "New desc"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				return fmt.Errorf("authentication required. Run: aphelion auth login")
			}

			serviceID := args[0]

			body := map[string]interface{}{}
			if cmd.Flags().Changed("price") {
				body["price"] = price
			}
			if cmd.Flags().Changed("description") {
				body["description"] = description
			}

			if len(body) == 0 {
				return fmt.Errorf("nothing to update. Specify --price and/or --description")
			}

			client := api.NewClient()

			spinner := utils.NewSpinner("Updating service...")
			spinner.Start()

			endpoint := fmt.Sprintf("/v2/services/%s", serviceID)
			err := client.Patch(endpoint, body, nil)
			spinner.Stop()

			if err != nil {
				return fmt.Errorf("failed to update service: %w", err)
			}

			utils.PrintSuccess("Service %s updated successfully", serviceID)
			return nil
		},
	}

	cmd.Flags().Float64Var(&price, "price", 0, "price per API call (e.g. 0.002)")
	cmd.Flags().StringVar(&description, "description", "", "service description")

	return cmd
}
