package invoke

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

var (
	inputJSON  string
	inputFile  string
	watch      bool
	outputFmt  string
)

// NewInvokeCmd creates the invoke command.
func NewInvokeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invoke <agent-name>",
		Short: "Invoke a deployed agent",
		Long: `Invoke a deployed agent with the given input.

The agent name can be a simple name for your own agents, or owner/agent-name
for marketplace agents published by other developers.`,
		Example: `  # Invoke your own agent
  aphelion invoke review-management-agent \
    --input '{"patient_name": "Jane", "contact": "+15551234567"}'

  # Invoke with input from a file
  aphelion invoke review-management-agent --input-file ./test-input.json

  # Invoke and stream logs while waiting
  aphelion invoke review-management-agent \
    --input '{"patient_name": "Jane", "contact": "+15551234567"}' --watch

  # Invoke a marketplace agent
  aphelion invoke exmplr/clinical-search \
    --input '{"query": "NSCLC phase 3 trials"}'

  # Output as table summary
  aphelion invoke review-management-agent \
    --input '{"patient_name": "Jane"}' --output table`,
		Args: cobra.ExactArgs(1),
		RunE: runInvoke,
	}

	cmd.Flags().StringVar(&inputJSON, "input", "", "inline JSON input for the agent")
	cmd.Flags().StringVar(&inputFile, "input-file", "", "path to JSON file containing agent input")
	cmd.Flags().BoolVar(&watch, "watch", false, "stream execution logs while waiting for result")
	cmd.Flags().StringVar(&outputFmt, "output", "json", "output format: json or table")

	return cmd
}

func runInvoke(cmd *cobra.Command, args []string) error {
	if !config.IsAuthenticated() {
		return fmt.Errorf("not authenticated.\nRun: aphelion auth login")
	}

	agentName := args[0]

	// Resolve input
	inputData, err := resolveInput()
	if err != nil {
		return err
	}

	// Validate input is valid JSON
	var inputMap map[string]interface{}
	if err := json.Unmarshal([]byte(inputData), &inputMap); err != nil {
		return fmt.Errorf("invalid JSON input: %w\nProvide valid JSON via --input or --input-file", err)
	}

	// Build endpoint based on agent name format
	endpoint := fmt.Sprintf("/v2/agents/%s/invoke", agentName)
	if strings.Contains(agentName, "/") {
		parts := strings.SplitN(agentName, "/", 2)
		endpoint = fmt.Sprintf("/v2/marketplace/%s/agents/%s/invoke", parts[0], parts[1])
	}

	client := api.NewClient()
	startTime := time.Now()

	// If --watch, poll logs in background while waiting
	if watch {
		logSince := time.Now().UTC().Format(time.RFC3339)
		done := make(chan struct{})
		go pollLogs(client, agentName, logSince, done)
		defer func() {
			close(done)
		}()
	}

	// Execute the invocation
	var result map[string]interface{}
	if err := client.Post(endpoint, inputMap, &result); err != nil {
		return err
	}

	elapsed := time.Since(startTime)

	// Print result
	fmt.Println()
	switch outputFmt {
	case "table":
		if err := utils.PrintOutput(result, "table"); err != nil {
			return err
		}
	default:
		encoded, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format result: %w", err)
		}
		fmt.Println(string(encoded))
	}

	fmt.Printf("\nExecution time: %s\n", formatDuration(elapsed))
	return nil
}

func resolveInput() (string, error) {
	if inputJSON != "" && inputFile != "" {
		return "", fmt.Errorf("cannot specify both --input and --input-file")
	}

	if inputJSON != "" {
		return inputJSON, nil
	}

	if inputFile != "" {
		data, err := os.ReadFile(inputFile)
		if err != nil {
			return "", fmt.Errorf("failed to read input file %q: %w", inputFile, err)
		}
		return string(data), nil
	}

	return "", fmt.Errorf("input is required.\nProvide inline JSON with --input or a file with --input-file\n\nExample:\n  aphelion invoke <agent> --input '{\"key\": \"value\"}'")
}

func pollLogs(client *api.Client, agentName string, since string, done <-chan struct{}) {
	// Strip owner prefix for log endpoint if marketplace agent
	logAgent := agentName
	if strings.Contains(agentName, "/") {
		parts := strings.SplitN(agentName, "/", 2)
		logAgent = parts[1]
	}

	endpoint := fmt.Sprintf("/v2/agents/%s/logs", logAgent)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			var logs []map[string]interface{}
			params := map[string]string{"since": since}
			if err := client.GetWithQuery(endpoint, params, &logs); err != nil {
				// Silently ignore log polling errors
				continue
			}
			for _, entry := range logs {
				if msg, ok := entry["message"]; ok {
					ts := ""
					if t, ok := entry["timestamp"]; ok {
						ts = fmt.Sprintf("[%v] ", t)
					}
					fmt.Printf("  %s%v\n", ts, msg)
				}
			}
			// Update since to latest timestamp
			if len(logs) > 0 {
				if last, ok := logs[len(logs)-1]["timestamp"]; ok {
					since = fmt.Sprintf("%v", last)
				}
			}
		}
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}
