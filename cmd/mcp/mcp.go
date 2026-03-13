package mcp

import "github.com/spf13/cobra"

// NewMCPCmd returns the parent mcp command.
func NewMCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP server for AI assistant integration",
		Long:  "Model Context Protocol (MCP) server that exposes Aphelion CLI functionality as tools for Claude Desktop, Claude.ai, and other MCP-compatible AI assistants.",
	}
	cmd.AddCommand(newServeCmd())
	cmd.AddCommand(newConfigCmd())
	return cmd
}
