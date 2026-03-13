package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newWhoamiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Show current authenticated user information",
		Long:  "Display the currently authenticated user's email, account ID, username, agent count, and API count.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				return fmt.Errorf("Not authenticated. Run: aphelion auth login")
			}

			client := api.NewClient()

			// Get user profile
			var profile map[string]interface{}
			if err := client.Get("/auth/test-profile", &profile); err != nil {
				return fmt.Errorf("Failed to fetch profile: %w", err)
			}

			email := stringFromMap(profile, "email")
			accountID := stringFromMap(profile, "id")
			username := stringFromMap(profile, "username")

			// Fall back to local config if API didn't return these
			if email == "" {
				email = config.GetUserEmail()
			}
			if accountID == "" {
				accountID = config.GetAccountID()
			}
			if username == "" {
				username = config.GetUsername()
			}

			// Get agent count
			agentCount := 0
			var agentsResp api.AgentsResponse
			if err := client.Get("/v2/agents", &agentsResp); err == nil {
				agentCount = agentsResp.Total
			}

			// Get API/service count
			apiCount := 0
			var servicesResp api.ServicesResponse
			if err := client.Get("/v2/services", &servicesResp); err == nil {
				apiCount = servicesResp.Total
			}

			fmt.Println()
			fmt.Printf("  Email:      %s\n", email)
			fmt.Printf("  Account ID: %s\n", accountID)
			fmt.Printf("  Username:   %s\n", username)
			fmt.Printf("  Agents:     %d\n", agentCount)
			fmt.Printf("  APIs:       %d\n", apiCount)
			fmt.Println()

			return nil
		},
	}

	return cmd
}

func stringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
