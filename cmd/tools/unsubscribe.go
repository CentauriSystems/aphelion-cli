package tools

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newUnsubscribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unsubscribe <tool-name>",
		Short: "Unsubscribe current agent from a tool",
		Long:  "Remove a tool subscription from the current project's agent.",
		Example: `  aphelion tools unsubscribe twilio
  aphelion tools unsubscribe sendgrid`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			toolName := args[0]

			if !config.IsProjectDir() {
				utils.PrintError("Not in an Aphelion project directory.")
				fmt.Println("Run this command from a directory with .aphelion/config.yaml")
				return fmt.Errorf("not in project directory")
			}

			agentID := config.GetAgentID()
			if agentID == "" {
				utils.PrintError("No agent ID found in project config.")
				return fmt.Errorf("no agent ID")
			}

			s := utils.NewSpinner(fmt.Sprintf("Unsubscribing from %s...", toolName))
			s.Start()

			client := api.NewClient()
			endpoint := fmt.Sprintf("/v2/agents/%s/tools/unsubscribe", agentID)
			err := client.DeleteWithBody(endpoint, map[string]string{"name": toolName})
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to unsubscribe from tool %q: %v", toolName, err)
				return err
			}

			// Update local project config
			cfg, err := config.LoadProjectConfig()
			if err != nil {
				utils.PrintError("Failed to read project config: %v", err)
				return err
			}
			if cfg != nil {
				updated := make([]string, 0, len(cfg.Tools.Subscribed))
				for _, t := range cfg.Tools.Subscribed {
					if t != toolName {
						updated = append(updated, t)
					}
				}
				cfg.Tools.Subscribed = updated
				if err := config.SaveProjectConfig(cfg); err != nil {
					utils.PrintError("Unsubscribed on server but failed to update local config: %v", err)
					return err
				}
			}

			utils.PrintSuccess("Unsubscribed from tool: %s", toolName)

			return nil
		},
	}

	return cmd
}
