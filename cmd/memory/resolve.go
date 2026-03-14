package memory

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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

// getAgentToken exchanges agent client_credentials for a short-lived agent JWT.
// Memory v2 endpoints require agent auth, not account auth.
func getAgentToken() (string, error) {
	clientID, clientSecret := config.GetAgentCredentials()
	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("no agent credentials found in project config.\nCreate an agent identity: aphelion agents create --name <name>")
	}

	apiURL := config.GetAPIUrl()
	payload := fmt.Sprintf(`{"grant_type":"client_credentials","client_id":"%s","client_secret":"%s"}`, clientID, clientSecret)

	req, err := http.NewRequest("POST", apiURL+"/auth/agent/token", strings.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("agent authentication failed (HTTP %d)", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse agent token response: %w", err)
	}

	return tokenResp.AccessToken, nil
}
