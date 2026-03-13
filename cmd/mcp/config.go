package mcp

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Print Claude Desktop MCP configuration",
		Long: `Prints the JSON configuration needed to add Aphelion as an MCP server in Claude Desktop.

Copy the output and paste it into your Claude Desktop settings (Settings > Developer > Edit Config).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(`{
  "mcpServers": {
    "aphelion": {
      "command": "aphelion",
      "args": ["mcp", "serve"],
      "env": {}
    }
  }
}`)
			return nil
		},
	}
}
