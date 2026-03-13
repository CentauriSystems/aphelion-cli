package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Exmplr-AI/aphelion-cli/cmd/agent"
	"github.com/Exmplr-AI/aphelion-cli/cmd/agents"
	"github.com/Exmplr-AI/aphelion-cli/cmd/analytics"
	"github.com/Exmplr-AI/aphelion-cli/cmd/auth"
	"github.com/Exmplr-AI/aphelion-cli/cmd/deploy"
	"github.com/Exmplr-AI/aphelion-cli/cmd/deployments"
	"github.com/Exmplr-AI/aphelion-cli/cmd/env"
	"github.com/Exmplr-AI/aphelion-cli/cmd/invoke"
	mcpCmd "github.com/Exmplr-AI/aphelion-cli/cmd/mcp"
	"github.com/Exmplr-AI/aphelion-cli/cmd/memory"
	"github.com/Exmplr-AI/aphelion-cli/cmd/registry"
	"github.com/Exmplr-AI/aphelion-cli/cmd/schedule"
	"github.com/Exmplr-AI/aphelion-cli/cmd/status"
	"github.com/Exmplr-AI/aphelion-cli/cmd/tools"
	authPkg "github.com/Exmplr-AI/aphelion-cli/pkg/auth"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

var (
	cfgFile   string
	apiURL    string
	output    string
	verbose   bool
)

var rootCmd = &cobra.Command{
	Use:   "aphelion",
	Short: "Aphelion Gateway CLI - Unified AI agent platform",
	Long: `Aphelion CLI provides command-line access to the Aphelion Gateway platform,
a unified AI agent platform with tool discovery and memory capabilities.

Use this CLI to manage authentication, register services, access memories,
and view analytics for your Aphelion Gateway account.`,
	Example: `  # Login to your account
  aphelion auth login

  # Initialize a new agent project
  aphelion agent init

  # Run an agent with cron scheduling
  aphelion agent run ./agent.py --cron "*/10 * * * *"

  # List available services
  aphelion registry list

  # Add service from OpenAPI spec
  aphelion registry add-openapi --file ./openapi.json

  # Describe a tool
  aphelion tools describe exmplr_core.search

  # Try executing a tool
  aphelion tools try --tool exmplr_core.search --params '{"q": "cancer"}'

  # Search your memories
  aphelion memory search "calculation"

  # View your usage analytics
  aphelion analytics user`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.aphelion/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "https://api.aphl.ai", "API base URL")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "output format (json|yaml|table)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	viper.BindPFlag("api-url", rootCmd.PersistentFlags().Lookup("api-url"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	rootCmd.AddCommand(auth.NewAuthCmd())
	rootCmd.AddCommand(agent.NewAgentCmd())
	rootCmd.AddCommand(agents.NewAgentsCmd())
	rootCmd.AddCommand(registry.NewRegistryCmd())
	rootCmd.AddCommand(memory.NewMemoryCmd())
	rootCmd.AddCommand(analytics.NewAnalyticsCmd())
	rootCmd.AddCommand(tools.NewToolsCmd())
	rootCmd.AddCommand(deploy.NewDeployCmd())
	rootCmd.AddCommand(deployments.NewDeploymentsCmd())
	rootCmd.AddCommand(env.NewEnvCmd())
	rootCmd.AddCommand(invoke.NewInvokeCmd())
	rootCmd.AddCommand(mcpCmd.NewMCPCmd())
	rootCmd.AddCommand(schedule.NewScheduleCmd())
	rootCmd.AddCommand(status.NewStatusCmd())
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newCompletionCmd())
	rootCmd.AddCommand(newWhoamiCmd())
	rootCmd.AddCommand(newOpenCmd())
	rootCmd.AddCommand(newQuickstartCmd())
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		configDir := filepath.Join(home, ".aphelion")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Printf("Warning: Could not create config directory: %v\n", err)
		}

		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.SetEnvPrefix("APHELION")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	}

	config.InitConfig()

	// Register the token refresh callback to avoid circular imports
	config.RefreshFunc = func(refreshToken string) (string, int, error) {
		resp, err := authPkg.RefreshAccessToken(refreshToken)
		if err != nil {
			return "", 0, err
		}
		return resp.AccessToken, resp.ExpiresIn, nil
	}
}