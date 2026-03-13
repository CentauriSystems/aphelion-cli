package agents

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newCreateCmd() *cobra.Command {
	var name string
	var description string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new agent identity",
		Long:  "Create a new agent identity on the Aphelion platform. Returns client credentials.",
		Example: `  aphelion agents create --name review-agent --description "Sends review requests"
  aphelion agents create --name data-pipeline --description "ETL pipeline agent"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			s := utils.NewSpinner("Creating agent identity...")
			s.Start()

			client := api.NewClient()
			req := api.CreateAgentRequest{
				Name:        name,
				Description: description,
			}

			var agent api.AgentIdentity
			err := client.Post("/v2/agents", req, &agent)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to create agent: %v", err)
				return err
			}

			utils.PrintSuccess("Agent created: %s", agent.Name)
			fmt.Println()

			bold := color.New(color.Bold)
			bold.Println("Agent Credentials")
			fmt.Println(strings.Repeat("-", 50))
			fmt.Printf("  Agent ID:      %s\n", agent.ID)
			fmt.Printf("  Client ID:     %s\n", agent.ClientID)
			fmt.Printf("  Client Secret: %s\n", agent.ClientSecret)
			fmt.Println(strings.Repeat("-", 50))
			fmt.Println()

			warning := color.New(color.FgYellow, color.Bold)
			warning.Println("WARNING: The client secret is shown ONCE. Save it now.")
			warning.Println("If you lose it, use: aphelion agents rotate-secret " + agent.Name)
			fmt.Println()

			// Offer to save to project config if in a project directory
			if config.IsProjectDir() {
				fmt.Print("Save credentials to .aphelion/config.yaml? [Y/n] ")
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))

				if answer == "" || answer == "y" || answer == "yes" {
					projCfg, err := config.LoadProjectConfig()
					if err != nil {
						utils.PrintError("Failed to load project config: %v", err)
						return nil
					}
					if projCfg == nil {
						projCfg = &config.ProjectConfig{}
					}
					projCfg.Agent.ID = agent.ID
					projCfg.Agent.ClientID = agent.ClientID
					projCfg.Agent.ClientSecret = agent.ClientSecret

					if err := config.SaveProjectConfig(projCfg); err != nil {
						utils.PrintError("Failed to save project config: %v", err)
						return nil
					}
					utils.PrintSuccess("Credentials saved to .aphelion/config.yaml")
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Agent name (required)")
	cmd.Flags().StringVar(&description, "description", "", "Agent description (required)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("description")

	return cmd
}
