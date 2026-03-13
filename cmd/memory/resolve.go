package memory

import (
	"fmt"

	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

// resolveAgentID determines the agent ID from the flag, project config, or returns an error.
func resolveAgentID(agentFlag string) (string, error) {
	if agentFlag != "" {
		return agentFlag, nil
	}
	if config.IsProjectDir() {
		if id := config.GetAgentID(); id != "" {
			return id, nil
		}
	}
	return "", fmt.Errorf("no agent specified. Use --agent flag or run from an agent project directory")
}
