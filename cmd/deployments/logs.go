package deployments

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

func newLogsCmd() *cobra.Command {
	var follow bool
	var limit int

	cmd := &cobra.Command{
		Use:   "logs <agent-name>",
		Short: "View logs from a deployed agent",
		Long:  "View recent logs from a deployed agent. Use --follow to stream logs in real time.",
		Args:  cobra.ExactArgs(1),
		Example: `  aphelion deployments logs my-agent
  aphelion deployments logs my-agent --follow
  aphelion deployments logs my-agent --limit 50`,
		RunE: func(cmd *cobra.Command, args []string) error {
			agentName := args[0]

			if !config.IsAuthenticated() {
				utils.PrintError("Not authenticated. Run: aphelion auth login")
				return fmt.Errorf("not authenticated")
			}

			client := api.NewClient()
			endpoint := fmt.Sprintf("/v2/agents/%s/logs", agentName)
			params := map[string]string{
				"limit": fmt.Sprintf("%d", limit),
			}

			if !follow {
				var resp api.LogsResponse
				err := client.GetWithQuery(endpoint, params, &resp)
				if err != nil {
					utils.PrintError("Failed to fetch logs: %v", err)
					return err
				}

				if len(resp.Logs) == 0 {
					fmt.Println("No logs found.")
					return nil
				}

				for _, entry := range resp.Logs {
					fmt.Printf("[%s] %s  %s\n", entry.Timestamp, entry.Level, entry.Message)
				}

				return nil
			}

			// Follow mode: poll every 2 seconds
			fmt.Printf("Streaming logs for %s (Ctrl+C to stop)...\n\n", agentName)
			lastTimestamp := ""

			for {
				queryParams := map[string]string{
					"limit": fmt.Sprintf("%d", limit),
				}
				if lastTimestamp != "" {
					queryParams["after"] = lastTimestamp
				}

				var resp api.LogsResponse
				err := client.GetWithQuery(endpoint, queryParams, &resp)
				if err != nil {
					utils.PrintError("Failed to fetch logs: %v", err)
					time.Sleep(2 * time.Second)
					continue
				}

				for _, entry := range resp.Logs {
					fmt.Printf("[%s] %s  %s\n", entry.Timestamp, entry.Level, entry.Message)
					lastTimestamp = entry.Timestamp
				}

				time.Sleep(2 * time.Second)
			}
		},
	}

	cmd.Flags().BoolVar(&follow, "follow", false, "Stream logs in real time")
	cmd.Flags().IntVar(&limit, "limit", 100, "Number of log lines to retrieve")

	return cmd
}
