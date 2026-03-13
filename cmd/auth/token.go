package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Print current bearer token",
		Long:  "Print the current access token to stdout for use in scripts and debugging",
		RunE: func(cmd *cobra.Command, args []string) error {
			token := config.GetAccessToken()
			if token == "" {
				return fmt.Errorf("not authenticated. Run: aphelion auth login")
			}

			fmt.Print(token)
			return nil
		},
	}

	return cmd
}
