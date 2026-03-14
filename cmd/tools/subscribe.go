package tools

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

type subscribeRequest struct {
	Name string `json:"name"`
}

func newSubscribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscribe <tool-name>",
		Short: "Subscribe current agent to a tool",
		Long:  "Subscribe the current project's agent to a tool from the marketplace.",
		Example: `  aphelion tools subscribe twilio
  aphelion tools subscribe sendgrid`,
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

			s := utils.NewSpinner(fmt.Sprintf("Subscribing to %s...", toolName))
			s.Start()

			client := api.NewClient()
			endpoint := fmt.Sprintf("/v2/agents/%s/tools/subscribe", agentID)
			var resp map[string]interface{}
			err := client.Post(endpoint, subscribeRequest{Name: toolName}, &resp)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to subscribe to tool %q: %v", toolName, err)
				return err
			}

			// Update local project config
			cfg, err := config.LoadProjectConfig()
			if err != nil {
				utils.PrintError("Failed to read project config: %v", err)
				return err
			}
			if cfg != nil {
				// Check if already in list
				found := false
				for _, t := range cfg.Tools.Subscribed {
					if t == toolName {
						found = true
						break
					}
				}
				if !found {
					cfg.Tools.Subscribed = append(cfg.Tools.Subscribed, toolName)
					if err := config.SaveProjectConfig(cfg); err != nil {
						utils.PrintError("Subscribed on server but failed to update local config: %v", err)
						return err
					}
				}
			}

			utils.PrintSuccess("Subscribed to tool: %s", toolName)
			fmt.Printf("Use in your agent: await tools.execute(\"%s.<operation>\", params)\n", toolName)
			fmt.Printf("Describe available operations: aphelion tools describe %s\n", toolName)

			return nil
		},
	}

	return cmd
}
