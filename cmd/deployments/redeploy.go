package deployments

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newRedeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redeploy <agent-name>",
		Short: "Redeploy an agent from current code",
		Long:  "Re-run the deployment process for an agent using the current code.",
		Args:  cobra.ExactArgs(1),
		Example: `  aphelion deployments redeploy my-agent`,
		RunE: func(cmd *cobra.Command, args []string) error {
			agentName := args[0]

			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			s := utils.NewSpinner("Redeploying agent...")
			s.Start()

			client := api.NewClient()
			endpoint := fmt.Sprintf("/v2/agents/%s/redeploy", agentName)
			var resp api.RedeployResponse
			err := client.Post(endpoint, nil, &resp)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to redeploy agent: %v", err)
				return err
			}

			utils.PrintSuccess("Agent redeployed successfully.")
			fmt.Printf("  Status:   %s\n", resp.Status)
			if resp.Endpoint != "" {
				fmt.Printf("  Endpoint: %s\n", resp.Endpoint)
			}
			if resp.Message != "" {
				fmt.Printf("  Message:  %s\n", resp.Message)
			}

			return nil
		},
	}

	return cmd
}
