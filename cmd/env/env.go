package env

import (
	"fmt"

	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
	"github.com/spf13/cobra"
)

// NewEnvCmd returns the env parent command.
func NewEnvCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Manage environment variables for deployed agents",
		Long: `Manage environment variables (secrets) for deployed agents.

Environment variables are stored server-side and injected at runtime.
Values are never displayed in plain text after being set.`,
	}

	cmd.AddCommand(newSetCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newPullCmd())
	cmd.AddCommand(newPushCmd())

	return cmd
}

// resolveAgentID determines the agent ID from the project config or returns an error.
func resolveAgentID(agentFlag string) (string, error) {
	if agentFlag != "" {
		return agentFlag, nil
	}
	if config.IsProjectDir() {
		if id := config.GetAgentID(); id != "" {
			return id, nil
		}
	}
	return "", fmt.Errorf("not in an agent project directory.\nRun from a project directory or use: aphelion env set --agent <name> KEY VALUE")
}

// resolveAgentName returns the project name from config, or the agent ID as fallback.
func resolveAgentName() string {
	cfg, err := config.LoadProjectConfig()
	if err == nil && cfg != nil && cfg.Name != "" {
		return cfg.Name
	}
	return "agent"
}
