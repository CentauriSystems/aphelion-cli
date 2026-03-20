package status

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

// agentStatusResponse holds the response from GET /v2/agents/{id}.
type agentStatusResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// memoryStatsResponse holds the response from GET /v2/agents/{id}/memory/stats.
type memoryStatsResponse struct {
	TotalEntries int `json:"total_entries"`
}

// memoryEntriesResponse holds the response from GET /v2/agents/{id}/memory/entries.
type memoryEntriesResponse struct {
	Entries []interface{} `json:"entries"`
	Total   int           `json:"total"`
}

// deploymentsListResponse holds the response from GET /v2/agents/{id}/deployments (plural).
type deploymentsListResponse struct {
	Deployments []api.DeploymentSummary `json:"deployments"`
	Total       int                     `json:"total"`
}

// toolSubscription represents a single tool subscription from the API.
type toolSubscription struct {
	ServiceName string `json:"service_name"`
	Name        string `json:"name"`
	Status      string `json:"status"`
}

// toolSubscriptionsResponse holds the response from GET /v2/agents/{id}/tools.
type toolSubscriptionsResponse struct {
	Tools []toolSubscription `json:"tools"`
	Total int                `json:"total"`
}

// executionStatsResponse holds execution statistics from the API.
type executionStatsResponse struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}

// envEntry holds a single env var from the API.
type envEntry struct {
	Key string `json:"key"`
}

// envKeysResponse holds the response from GET /v2/agents/{id}/env.
type envKeysResponse struct {
	Env []envEntry `json:"env"`
}

// NewStatusCmd returns the top-level status command.
func NewStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show project status at a glance",
		Long: `Show a comprehensive status dashboard for the current agent project.

Run this command from inside an agent project directory (containing .aphelion/config.yaml).
It displays agent identity, deployment status, tool subscriptions, scheduling,
environment variables, memory stats, and execution history.`,
		Example: `  aphelion status`,
		RunE:    runStatus,
	}

	return cmd
}

func runStatus(cmd *cobra.Command, args []string) error {
	if !config.IsProjectDir() {
		utils.PrintError("Not in an Aphelion project directory.")
		fmt.Println("Run this command from a directory with .aphelion/config.yaml")
		fmt.Println("Or initialize a project with: aphelion agent init")
		return fmt.Errorf("not in project directory")
	}

	projCfg, err := config.LoadProjectConfig()
	if err != nil {
		utils.PrintError("Failed to load project config: %v", err)
		return err
	}
	if projCfg == nil {
		utils.PrintError("Failed to load project config.")
		return fmt.Errorf("project config is nil")
	}

	agentName := projCfg.Name
	agentID := projCfg.Agent.ID

	if agentID == "" {
		utils.PrintError("No agent ID found in project config.")
		fmt.Println("Create an agent identity with: aphelion agents create --name <name> --description <desc>")
		return fmt.Errorf("no agent ID configured")
	}

	client := api.NewClient()

	// Fetch agent status (best-effort)
	var agentStatus agentStatusResponse
	agentErr := client.Get(fmt.Sprintf("/v2/agents/%s", agentID), &agentStatus)

	// Fetch deployment info — try singular endpoint first, then plural
	var deployment api.DeploymentInfo
	deployErr := client.Get(fmt.Sprintf("/v2/agents/%s/deployment", agentID), &deployment)
	if deployErr != nil || deployment.Status == "" {
		// Try the plural /deployments endpoint (used by the deploy command)
		var deploys deploymentsListResponse
		if listErr := client.Get(fmt.Sprintf("/v2/agents/%s/deployments", agentID), &deploys); listErr == nil && len(deploys.Deployments) > 0 {
			latest := deploys.Deployments[0]
			deployment.Status = latest.Status
			deployment.Endpoint = latest.Endpoint
			deployment.Region = latest.Region
			deployment.LastDeployed = latest.LastDeployed
			deployErr = nil
		}
	}

	// If API didn't return deployment info but local config says deployed, trust local config
	if (deployErr != nil || deployment.Status == "") &&
		(projCfg.Deployment.Status == "deployed" || projCfg.Deployment.Status == "active") &&
		projCfg.Deployment.Endpoint != "" {
		deployment.Status = projCfg.Deployment.Status
		deployment.Endpoint = projCfg.Deployment.Endpoint
		deployment.Region = projCfg.Deployment.Region
		deployment.LastDeployed = projCfg.Deployment.LastDeployed
		deployErr = nil
	}

	// Fetch memory stats — try /memory/stats first, then fall back to /memory/entries
	var memoryCount int
	var memStats memoryStatsResponse
	if statsErr := client.Get(fmt.Sprintf("/v2/agents/%s/memory/stats", agentID), &memStats); statsErr == nil && memStats.TotalEntries > 0 {
		memoryCount = memStats.TotalEntries
	} else {
		// Fall back to /memory/entries which returns a list we can count
		var memEntries memoryEntriesResponse
		if entriesErr := client.Get(fmt.Sprintf("/v2/agents/%s/memory/entries", agentID), &memEntries); entriesErr == nil {
			if memEntries.Total > 0 {
				memoryCount = memEntries.Total
			} else {
				memoryCount = len(memEntries.Entries)
			}
		}
	}

	// Fetch tool subscriptions from API
	var apiTools []string
	var toolSubs toolSubscriptionsResponse
	if toolErr := client.Get(fmt.Sprintf("/v2/agents/%s/tools", agentID), &toolSubs); toolErr == nil && len(toolSubs.Tools) > 0 {
		for _, t := range toolSubs.Tools {
			name := t.ServiceName
			if name == "" {
				name = t.Name
			}
			if name != "" {
				apiTools = append(apiTools, name)
			}
		}
	}
	// Merge with local config tools (API takes precedence, but include local ones not in API)
	toolSet := make(map[string]bool)
	for _, t := range apiTools {
		toolSet[t] = true
	}
	for _, t := range projCfg.Tools.Subscribed {
		toolSet[t] = true
	}
	var allTools []string
	// Preserve API order first, then local-only tools
	for _, t := range apiTools {
		allTools = append(allTools, t)
	}
	for _, t := range projCfg.Tools.Subscribed {
		if !containsIgnoreCase(apiTools, t) {
			allTools = append(allTools, t)
		}
	}

	// Fetch recent executions (ignore errors)
	var execResp api.ExecutionsResponse
	_ = client.GetWithQuery(fmt.Sprintf("/v2/agents/%s/executions", agentID), map[string]string{"limit": "1"}, &execResp)

	// Fetch execution stats (ignore errors)
	var execStats executionStatsResponse
	_ = client.Get(fmt.Sprintf("/v2/agents/%s/executions/stats", agentID), &execStats)

	// If execution stats came back empty but we have executions response with a total, use that
	if execStats.Total == 0 && execResp.Total > 0 {
		execStats.Total = execResp.Total
	}

	// Fetch env var keys (ignore errors)
	var envKeys envKeysResponse
	_ = client.Get(fmt.Sprintf("/v2/agents/%s/env", agentID), &envKeys)

	// Build the dashboard
	line := strings.Repeat("─", 50)

	fmt.Println()
	fmt.Println(line)
	fmt.Printf("  %s\n", agentName)
	fmt.Println(line)

	// Agent ID
	fmt.Printf("  Agent ID:     %s\n", agentID)

	// Status
	if agentErr != nil && deployErr != nil {
		fmt.Printf("  Status:       unknown (could not reach API)\n")
	} else if deployment.Status == "active" || deployment.Status == "deployed" {
		fmt.Printf("  Status:       deployed ✓\n")
	} else if deployErr == nil && deployment.Status != "" {
		fmt.Printf("  Status:       %s\n", deployment.Status)
	} else {
		fmt.Printf("  Status:       not deployed\n")
	}

	// Endpoint
	if deployment.Endpoint != "" {
		fmt.Printf("  Endpoint:     %s\n", deployment.Endpoint)
	} else {
		fmt.Printf("  Endpoint:     —\n")
	}

	// Region
	region := deployment.Region
	if region == "" {
		region = projCfg.Deployment.Region
	}
	if region == "" {
		region = "us-central1"
	}
	fmt.Printf("  Region:       %s\n", region)

	// Last deploy
	if deployment.LastDeployed != "" {
		fmt.Printf("  Last deploy:  %s\n", relativeTime(deployment.LastDeployed))
	} else if projCfg.Deployment.LastDeployed != "" {
		fmt.Printf("  Last deploy:  %s\n", relativeTime(projCfg.Deployment.LastDeployed))
	} else {
		fmt.Printf("  Last deploy:  —\n")
	}

	fmt.Println()

	// Tools
	if len(allTools) > 0 {
		toolParts := make([]string, len(allTools))
		for i, t := range allTools {
			toolParts[i] = t + " ✓"
		}
		fmt.Printf("  Tools:        %s\n", strings.Join(toolParts, "  "))
	} else {
		fmt.Printf("  Tools:        none\n")
	}

	// Schedule
	if projCfg.Schedule.Cron != "" {
		enabledStr := "disabled"
		if projCfg.Schedule.Enabled {
			enabledStr = "enabled"
		}
		fmt.Printf("  Schedule:     %s (%s)\n", projCfg.Schedule.Cron, enabledStr)
	} else {
		fmt.Printf("  Schedule:     none\n")
	}

	// Env vars
	var envKeyNames []string
	for _, e := range envKeys.Env {
		envKeyNames = append(envKeyNames, e.Key)
	}
	if len(envKeyNames) > 0 {
		fmt.Printf("  Env vars:     %s\n", strings.Join(envKeyNames, ", "))
	} else {
		fmt.Printf("  Env vars:     none set\n")
	}

	fmt.Println()

	// Memory
	fmt.Printf("  Memory:       %d entries\n", memoryCount)

	// Executions
	if execStats.Total > 0 {
		rate := float64(0)
		if execStats.Total > 0 {
			rate = float64(execStats.Completed) / float64(execStats.Total) * 100
		}
		fmt.Printf("  Executions:   %d total  |  %d success (%.1f%%)  |  %d failed\n",
			execStats.Total, execStats.Completed, rate, execStats.Failed)
	} else {
		fmt.Printf("  Executions:   0 total\n")
	}

	// Last run
	if len(execResp.Executions) > 0 {
		last := execResp.Executions[0]
		fmt.Printf("  Last run:     %s (%s, %s)\n",
			relativeTime(last.Timestamp), last.Status, last.Duration)
	} else {
		fmt.Printf("  Last run:     —\n")
	}

	fmt.Println(line)
	fmt.Printf("  Console: https://beta.console.aphl.ai/agents/%s\n", agentName)
	fmt.Println()

	return nil
}

// contains checks if a string slice contains a value.
func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

func containsIgnoreCase(slice []string, val string) bool {
	lower := strings.ToLower(val)
	for _, s := range slice {
		if strings.ToLower(s) == lower {
			return true
		}
	}
	return false
}

// relativeTime converts a timestamp string to a human-readable relative time.
func relativeTime(timestamp string) string {
	if timestamp == "" {
		return "—"
	}

	// Try common formats
	var t time.Time
	var err error
	for _, layout := range []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	} {
		t, err = time.Parse(layout, timestamp)
		if err == nil {
			break
		}
	}
	if err != nil {
		return timestamp
	}

	diff := time.Since(t)
	if diff < 0 {
		return "just now"
	}

	seconds := int(math.Floor(diff.Seconds()))
	minutes := int(math.Floor(diff.Minutes()))
	hours := int(math.Floor(diff.Hours()))
	days := int(math.Floor(diff.Hours() / 24))

	switch {
	case seconds < 60:
		return "just now"
	case minutes == 1:
		return "1 minute ago"
	case minutes < 60:
		return fmt.Sprintf("%d minutes ago", minutes)
	case hours == 1:
		return "1 hour ago"
	case hours < 24:
		return fmt.Sprintf("%d hours ago", hours)
	case days == 1:
		return "1 day ago"
	case days < 30:
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("Jan 2, 2006")
	}
}
