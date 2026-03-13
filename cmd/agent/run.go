package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

var (
	cronSchedule string
	daemon       bool
	verbose      bool
	input        string
	inputFile    string
	watch        bool
)

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [agent-file]",
		Short: "Run an agent",
		Long:  "Execute an agent script with optional cron scheduling",
		Args:  cobra.ExactArgs(1),
		RunE:  runAgent,
	}

	cmd.Flags().StringVar(&cronSchedule, "cron", "", "Cron schedule for agent execution (e.g., '*/10 * * * *')")
	cmd.Flags().BoolVarP(&daemon, "daemon", "d", false, "Run agent as daemon")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	cmd.Flags().StringVar(&input, "input", "", "JSON input to pass to agent")
	cmd.Flags().StringVar(&inputFile, "input-file", "", "Path to JSON input file")
	cmd.Flags().BoolVar(&watch, "watch", false, "Stream logs in real time")

	return cmd
}

// resolveInput returns the JSON input string from --input or --input-file flags.
// Returns empty string if neither is set. Returns error if input is invalid.
func resolveInput() (string, error) {
	if input != "" && inputFile != "" {
		return "", fmt.Errorf("cannot specify both --input and --input-file")
	}

	if input != "" {
		// Validate it's valid JSON
		if !json.Valid([]byte(input)) {
			return "", fmt.Errorf("--input value is not valid JSON")
		}
		return input, nil
	}

	if inputFile != "" {
		data, err := os.ReadFile(inputFile)
		if err != nil {
			return "", fmt.Errorf("failed to read input file: %w", err)
		}
		if !json.Valid(data) {
			return "", fmt.Errorf("input file does not contain valid JSON")
		}
		return string(data), nil
	}

	return "", nil
}

// exchangeAgentToken attempts to exchange agent credentials for a short-lived JWT.
// Returns the token string or empty string if credentials are not available.
func exchangeAgentToken() string {
	if !config.IsProjectDir() {
		return ""
	}

	clientID, clientSecret := config.GetAgentCredentials()
	if clientID == "" || clientSecret == "" {
		return ""
	}

	client := api.NewClient()
	tokenReq := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"grant_type":    "client_credentials",
	}
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := client.Post("/auth/agent/token", tokenReq, &tokenResp); err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Warning: failed to exchange agent credentials for token: %v\n", err)
		}
		return ""
	}
	return tokenResp.AccessToken
}

// buildAgentEnv returns the environment variables to inject into the agent subprocess.
func buildAgentEnv(agentToken, inputJSON string) []string {
	env := append(os.Environ(),
		"APHELION_API_URL="+config.GetAPIUrl(),
	)

	if agentID := config.GetAgentID(); agentID != "" {
		env = append(env, "APHELION_AGENT_ID="+agentID)
	}

	if agentToken != "" {
		env = append(env, "APHELION_API_TOKEN="+agentToken)
	}

	// Generate a session ID
	sessionID := fmt.Sprintf("ses_%d", time.Now().UnixNano())
	env = append(env, "APHELION_SESSION_ID="+sessionID)

	if inputJSON != "" {
		env = append(env, "APHELION_INPUT="+inputJSON)
	}

	return env
}

func runAgent(cmd *cobra.Command, args []string) error {
	agentFile := args[0]

	// Validate agent file exists
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		return fmt.Errorf("agent file not found: %s\nCheck the file path and try again", agentFile)
	}

	// Make file path absolute
	absPath, err := filepath.Abs(agentFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Resolve input
	inputJSON, err := resolveInput()
	if err != nil {
		return err
	}

	if cronSchedule != "" {
		return runWithCron(absPath, inputJSON)
	}

	if daemon {
		return runAsDaemon(absPath, inputJSON)
	}

	return runOnce(absPath, inputJSON)
}

func runOnce(agentFile string, inputJSON string) error {
	if verbose {
		fmt.Printf("Running agent: %s\n", agentFile)
	}

	// Exchange agent credentials for JWT
	agentToken := exchangeAgentToken()

	agentCmd := createAgentCommand(agentFile)
	agentCmd.Env = buildAgentEnv(agentToken, inputJSON)
	agentCmd.Stdout = os.Stdout
	agentCmd.Stderr = os.Stderr

	startTime := time.Now()
	err := agentCmd.Run()
	duration := time.Since(startTime)

	if err != nil {
		return fmt.Errorf("agent execution failed after %s: %w", duration.Round(time.Millisecond), err)
	}

	utils.PrintSuccess("Agent completed in %s", duration.Round(time.Millisecond))
	return nil
}

func runWithCron(agentFile string, inputJSON string) error {
	fmt.Printf("Scheduling agent with cron: %s\n", cronSchedule)
	fmt.Printf("Agent file: %s\n", agentFile)

	c := cron.New()

	_, err := c.AddFunc(cronSchedule, func() {
		if verbose {
			fmt.Printf("[%s] Running scheduled agent execution\n", time.Now().Format("2006-01-02 15:04:05"))
		}

		agentToken := exchangeAgentToken()
		agentCmd := createAgentCommand(agentFile)
		agentCmd.Env = buildAgentEnv(agentToken, inputJSON)

		if verbose {
			agentCmd.Stdout = os.Stdout
			agentCmd.Stderr = os.Stderr
		}

		startTime := time.Now()
		if err := agentCmd.Run(); err != nil {
			fmt.Printf("Agent execution failed: %v\n", err)
		} else {
			duration := time.Since(startTime)
			if verbose {
				utils.PrintSuccess("Agent execution completed in %s", duration.Round(time.Millisecond))
			}
		}
	})

	if err != nil {
		return fmt.Errorf("invalid cron schedule: %w", err)
	}

	c.Start()
	defer c.Stop()

	fmt.Println("Cron scheduler started. Press Ctrl+C to stop.")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nStopping cron scheduler...")
	return nil
}

func runAsDaemon(agentFile string, inputJSON string) error {
	fmt.Printf("Running agent as daemon: %s\n", agentFile)

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run agent in a goroutine
	done := make(chan error, 1)
	go func() {
		agentToken := exchangeAgentToken()
		agentCmd := createAgentCommand(agentFile)
		agentCmd.Env = buildAgentEnv(agentToken, inputJSON)

		if verbose {
			agentCmd.Stdout = os.Stdout
			agentCmd.Stderr = os.Stderr
		}

		done <- agentCmd.Run()
	}()

	// Wait for either completion or signal
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("agent execution failed: %w", err)
		}
		return nil
	case sig := <-sigChan:
		fmt.Printf("\nReceived signal %v, stopping agent...\n", sig)
		return nil
	}
}

func createAgentCommand(agentFile string) *exec.Cmd {
	// Determine how to run the agent based on file extension
	ext := strings.ToLower(filepath.Ext(agentFile))

	switch ext {
	case ".py":
		return exec.Command("python3", agentFile)
	case ".js":
		return exec.Command("node", agentFile)
	case ".go":
		return exec.Command("go", "run", agentFile)
	default:
		// Try to make it executable and run directly
		os.Chmod(agentFile, 0755)
		return exec.Command(agentFile)
	}
}
