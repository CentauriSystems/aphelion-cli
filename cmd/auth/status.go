package auth

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current authentication status",
		Long:  "Display the current authentication context, token validity, and user information",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.GetConfig()

			if !config.IsAuthenticated() {
				fmt.Println("Not authenticated.")
				fmt.Println("")
				fmt.Println("Run: aphelion auth login")
				return nil
			}

			email := config.GetUserEmail()
			accountID := config.GetAccountID()

			fmt.Println("Authenticated")
			fmt.Println("")
			fmt.Printf("  Context:    human (Auth0 PKCE)\n")

			if email != "" {
				fmt.Printf("  Email:      %s\n", email)
			}
			if accountID != "" {
				fmt.Printf("  Account ID: %s\n", accountID)
			}

			// Token expiry
			if cfg.Auth.ExpiresAt != "" {
				expiresAt, err := time.Parse(time.RFC3339, cfg.Auth.ExpiresAt)
				if err == nil {
					if time.Now().Before(expiresAt) {
						remaining := time.Until(expiresAt).Truncate(time.Second)
						fmt.Printf("  Token:      valid (expires in %s)\n", remaining)
					} else {
						if cfg.Auth.RefreshToken != "" {
							fmt.Printf("  Token:      expired (refresh token available)\n")
						} else {
							fmt.Printf("  Token:      expired\n")
						}
					}
					fmt.Printf("  Expires at: %s\n", expiresAt.Local().Format("2006-01-02 15:04:05 MST"))
				}
			} else {
				fmt.Printf("  Token:      valid (no expiry set)\n")
			}

			return nil
		},
	}

	return cmd
}
