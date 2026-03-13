package auth

import (
	"github.com/spf13/cobra"
)

func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
		Long:  "Manage authentication with the Aphelion Gateway API",
	}

	cmd.AddCommand(newLoginCmd())
	cmd.AddCommand(newLogoutCmd())
	cmd.AddCommand(newProfileCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newTokenCmd())

	return cmd
}