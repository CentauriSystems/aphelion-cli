package tools

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

type subscribeRequest struct {
	Name string `json:"name"`
}

type serviceSummary struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type servicesResponse struct {
	Services []serviceSummary `json:"services"`
}

// fuzzyMatchServices returns service names that contain the query as a
// case-insensitive substring.
func fuzzyMatchServices(query string, services []serviceSummary) []string {
	lower := strings.ToLower(query)
	var matches []string
	for _, svc := range services {
		if strings.Contains(strings.ToLower(svc.Name), lower) {
			matches = append(matches, svc.Name)
		}
	}
	return matches
}

func newSubscribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscribe <tool-name>",
		Short: "Subscribe current agent to a tool",
		Long:  "Subscribe the current project's agent to a tool from the marketplace.",
		Example: `  aphelion tools subscribe twilio
  aphelion tools subscribe sendgrid
  aphelion tools subscribe calculator`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			toolName := args[0]

			if !config.IsProjectDir() {
				utils.PrintError("Not in an Aphelion project directory.")
				fmt.Println("Run this command from a directory with .aphelion/config.yaml")
				fmt.Println("Or create a new project: aphelion agent init")
				return fmt.Errorf("not in project directory")
			}

			agentID := config.GetAgentID()
			if agentID == "" {
				utils.PrintError("No agent ID found in project config.")
				fmt.Println("Create an agent identity: aphelion agents create --name <name>")
				return fmt.Errorf("no agent ID")
			}

			client := api.NewClient()
			endpoint := fmt.Sprintf("/v2/agents/%s/tools/subscribe", agentID)

			// Try subscribing with the exact name the user provided.
			s := utils.NewSpinner(fmt.Sprintf("Subscribing to %s...", toolName))
			s.Start()

			var resp map[string]interface{}
			err := client.Post(endpoint, subscribeRequest{Name: toolName}, &resp)
			s.Stop()

			resolvedName := toolName

			if err != nil {
				// Exact name failed — attempt fuzzy match against marketplace.
				s2 := utils.NewSpinner("Searching marketplace for matching services...")
				s2.Start()

				var svcResp servicesResponse
				fetchErr := client.Get("/services/summary", &svcResp)
				s2.Stop()

				if fetchErr != nil {
					// Cannot fetch marketplace; report the original error.
					utils.PrintError("Failed to subscribe to tool %q: %v", toolName, err)
					fmt.Println("Could not search marketplace for alternatives.")
					return err
				}

				matches := fuzzyMatchServices(toolName, svcResp.Services)

				switch len(matches) {
				case 0:
					utils.PrintError("Service %q not found in the marketplace.", toolName)
					fmt.Println("Browse available tools: aphelion tools marketplace")
					fmt.Println("Search for tools:      aphelion tools search <query>")
					return fmt.Errorf("service not found: %s", toolName)

				case 1:
					// Exactly one match — auto-subscribe with the correct name.
					resolvedName = matches[0]
					fmt.Printf("Matched %q to service %q\n", toolName, resolvedName)

					s3 := utils.NewSpinner(fmt.Sprintf("Subscribing to %s...", resolvedName))
					s3.Start()

					var retryResp map[string]interface{}
					retryErr := client.Post(endpoint, subscribeRequest{Name: resolvedName}, &retryResp)
					s3.Stop()

					if retryErr != nil {
						utils.PrintError("Failed to subscribe to %q: %v", resolvedName, retryErr)
						return retryErr
					}

				default:
					// Multiple matches — let the user pick.
					utils.PrintError("Service %q is ambiguous. Did you mean one of these?", toolName)
					fmt.Println()
					for _, m := range matches {
						fmt.Printf("  aphelion tools subscribe %q\n", m)
					}
					fmt.Println()
					fmt.Println("Re-run with the exact service name from the list above.")
					return fmt.Errorf("ambiguous service name: %s", toolName)
				}
			}

			// Update local project config with the resolved (exact) service name.
			cfg, err := config.LoadProjectConfig()
			if err != nil {
				utils.PrintError("Failed to read project config: %v", err)
				return err
			}
			if cfg != nil {
				found := false
				for _, t := range cfg.Tools.Subscribed {
					if t == resolvedName {
						found = true
						break
					}
				}
				if !found {
					cfg.Tools.Subscribed = append(cfg.Tools.Subscribed, resolvedName)
					if err := config.SaveProjectConfig(cfg); err != nil {
						utils.PrintError("Subscribed on server but failed to update local config: %v", err)
						return err
					}
				}
			}

			utils.PrintSuccess("Subscribed to tool: %s", resolvedName)
			fmt.Printf("Use in your agent: await tools.execute(\"%s.<operation>\", params)\n", resolvedName)
			fmt.Printf("Describe available operations: aphelion tools describe %s\n", resolvedName)

			return nil
		},
	}

	return cmd
}
