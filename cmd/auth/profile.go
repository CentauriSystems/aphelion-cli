package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
)

func newProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Show user profile information",
		Long:  "Display the current user's profile information",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				return fmt.Errorf("not authenticated. Please run 'aphelion auth login' first")
			}

			client := api.NewClient()
			
			var profile api.User
			if err := client.Get("/auth/profile", &profile); err != nil {
				return fmt.Errorf("failed to get profile: %w", err)
			}

			data := map[string]interface{}{
				"ID":         profile.ID,
				"Username":   profile.Username,
				"Email":      profile.Email,
			}

			return utils.PrintOutput(data, config.GetOutputFormat())
		},
	}

	return cmd
}