package env

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newPushCmd() *cobra.Command {
	var (
		agentFlag string
		yes       bool
	)

	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push local .env variables to deployed agent",
		Long: `Read local .env file and push all KEY=VALUE pairs to the deployed agent's environment.
Prompts for confirmation before pushing.`,
		Example: `  # Push .env to deployed agent
  aphelion env push

  # Push without confirmation
  aphelion env push --yes

  # Push for a specific agent
  aphelion env push --agent review-management-agent`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				return fmt.Errorf("session expired. Run: aphelion auth login")
			}

			agentID, err := resolveAgentID(agentFlag)
			if err != nil {
				return err
			}

			// Parse .env file
			envVars, err := parseEnvFile(".env")
			if err != nil {
				return err
			}

			if len(envVars) == 0 {
				utils.PrintInfo("No environment variables found in .env")
				return nil
			}

			// Show what will be set
			fmt.Println("The following environment variables will be set:")
			for key := range envVars {
				fmt.Printf("  %s\n", key)
			}
			fmt.Println()

			// Prompt confirmation unless --yes
			if !yes {
				fmt.Printf("Push %d environment variables to deployed agent? [y/N] ", len(envVars))
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer != "y" && answer != "yes" {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			client := api.NewClient()
			var failed []string

			for key, value := range envVars {
				endpoint := fmt.Sprintf("/v2/agents/%s/env/%s", agentID, key)
				body := map[string]string{
					"value": value,
				}
				if err := client.Put(endpoint, body, nil); err != nil {
					failed = append(failed, key)
					utils.PrintError("Failed to set %s: %v", key, err)
				}
			}

			succeeded := len(envVars) - len(failed)
			if succeeded > 0 {
				utils.PrintSuccess("Pushed %d environment variables to deployed agent", succeeded)
			}
			if len(failed) > 0 {
				return fmt.Errorf("%d environment variables failed to push", len(failed))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&agentFlag, "agent", "", "agent name or ID")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation prompt")

	return cmd
}

// parseEnvFile reads a .env file and returns key-value pairs.
// Skips empty lines and lines starting with #.
func parseEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no .env file found in current directory.\nCreate a .env file with KEY=VALUE pairs, one per line")
		}
		return nil, fmt.Errorf("failed to read .env file: %w", err)
	}
	defer file.Close()

	envVars := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first = sign
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes if present
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		if key != "" {
			envVars[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading .env file: %w", err)
	}

	return envVars, nil
}
