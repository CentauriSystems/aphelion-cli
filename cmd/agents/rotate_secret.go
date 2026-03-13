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

func newRotateSecretCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotate-secret <name-or-id>",
		Short: "Rotate an agent's client secret",
		Long:  "Rotate the client secret for an agent identity. The current secret is immediately invalidated.",
		Example: `  aphelion agents rotate-secret review-agent
  aphelion agents rotate-secret agt_KFnG9Ad2r9LsCrJdLu9k`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			nameOrID := args[0]

			fmt.Print("This will immediately invalidate the current secret. Continue? [y/N] ")
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(answer)

			if answer != "y" && answer != "Y" {
				fmt.Println("Aborted.")
				return nil
			}

			s := utils.NewSpinner("Rotating secret...")
			s.Start()

			client := api.NewClient()
			var resp api.RotateSecretResponse
			err := client.Post(fmt.Sprintf("/v2/agents/%s/rotate-secret", nameOrID), nil, &resp)
			s.Stop()

			if err != nil {
				utils.PrintError("Failed to rotate secret for agent %q: %v", nameOrID, err)
				return err
			}

			utils.PrintSuccess("Secret rotated for agent %q", nameOrID)
			fmt.Println()

			bold := color.New(color.Bold)
			bold.Println("New Credentials")
			fmt.Println(strings.Repeat("-", 50))
			fmt.Printf("  Client ID:     %s\n", resp.ClientID)
			fmt.Printf("  Client Secret: %s\n", resp.ClientSecret)
			fmt.Println(strings.Repeat("-", 50))
			fmt.Println()

			warning := color.New(color.FgYellow, color.Bold)
			warning.Println("WARNING: The client secret is shown ONCE. Save it now.")
			fmt.Println()

			// If in a project directory and agent matches, update config
			if config.IsProjectDir() {
				projCfg, err := config.LoadProjectConfig()
				if err == nil && projCfg != nil && (projCfg.Agent.ID == nameOrID || projCfg.Name == nameOrID) {
					projCfg.Agent.ClientID = resp.ClientID
					projCfg.Agent.ClientSecret = resp.ClientSecret
					if err := config.SaveProjectConfig(projCfg); err != nil {
						utils.PrintError("Failed to update project config: %v", err)
					} else {
						utils.PrintSuccess("Updated .aphelion/config.yaml with new secret")
					}
				}
			}

			return nil
		},
	}

	return cmd
}
