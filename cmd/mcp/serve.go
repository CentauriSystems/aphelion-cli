package mcp

import (
	"github.com/spf13/cobra"

	mcpServer "github.com/Exmplr-AI/aphelion-cli/pkg/mcp"
)

func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start MCP server for AI assistant integration",
		Long: `Starts an MCP server on stdio for use with Claude Desktop, Claude.ai, and other MCP-compatible AI assistants.

The server communicates via JSON-RPC 2.0 over stdin/stdout. It exposes Aphelion CLI
functionality as MCP tools that AI assistants can call directly.

To configure Claude Desktop, run: aphelion mcp config`,
		Example: `  # Start the MCP server (typically called by Claude Desktop, not directly)
  aphelion mcp serve

  # Print Claude Desktop configuration
  aphelion mcp config`,
		RunE: func(cmd *cobra.Command, args []string) error {
			server := mcpServer.NewServer()
			return server.Run()
		},
	}
}
