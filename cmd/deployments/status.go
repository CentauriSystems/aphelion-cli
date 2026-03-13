package deployments

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show deployment status for current project",
		Long:  "Show the deployment status of the agent in the current project directory.",
		Example: `  aphelion deployments status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

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
			if projCfg == nil || projCfg.Agent.ID == "" {
				utils.PrintError("No agent ID found in project config.")
				fmt.Println("Create an agent identity with: aphelion agents create --name <name> --description <desc>")
				return fmt.Errorf("no agent ID configured")
			}

			s := utils.NewSpinner("Fetching deployment status...")
			s.Start()

			client := api.NewClient()
			var status api.DeploymentStatus
			endpoint := fmt.Sprintf("/v2/agents/%s/deployment", projCfg.Agent.ID)
			err = client.Get(endpoint, &status)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to get deployment status: %v", err)
				return err
			}

			outputFormat := config.GetOutputFormat()
			if f, _ := cmd.Flags().GetString("output"); f != "" {
				outputFormat = f
			}

			if outputFormat == "json" || outputFormat == "yaml" {
				return utils.PrintOutput(status, outputFormat)
			}

			fmt.Println()
			fmt.Printf("  Agent:           %s\n", status.AgentName)
			fmt.Printf("  Agent ID:        %s\n", status.AgentID)
			fmt.Printf("  Status:          %s\n", status.Status)
			fmt.Printf("  Endpoint:        %s\n", status.Endpoint)
			fmt.Printf("  Region:          %s\n", status.Region)
			fmt.Printf("  Version:         %s\n", status.Version)
			fmt.Printf("  Last Deployed:   %s\n", status.LastDeployed)
			fmt.Printf("  Execution Count: %d\n", status.ExecutionCount)
			fmt.Println()

			return nil
		},
	}

	return cmd
}
