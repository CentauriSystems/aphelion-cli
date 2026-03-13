package memory

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
)

func newSearchCmd() *cobra.Command {
	var (
		limit     int
		threshold float64
		agentFlag string
	)

	cmd := &cobra.Command{
		Use:   "search [QUERY]",
		Short: "Search through memories",
		Long:  "Search through agent memories using semantic similarity",
		Args:  cobra.ExactArgs(1),
		Example: `  # Search for memories about calculations
  aphelion memory search "calculation"

  # Search with custom threshold and limit
  aphelion memory search "weather" --threshold 0.8 --limit 5

  # Search a specific agent's memory
  aphelion memory search "patient contact history" --agent review-management-agent`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				return fmt.Errorf("session expired. Run: aphelion auth login")
			}

			query := args[0]
			client := api.NewClient()

			params := map[string]string{
				"q":         query,
				"limit":     strconv.Itoa(limit),
				"threshold": fmt.Sprintf("%.2f", threshold),
			}

			// Determine endpoint based on agent context
			var endpoint string
			agentID, agentErr := resolveAgentID(agentFlag)
			if agentErr == nil && agentID != "" {
				// Agent-scoped search
				endpoint = fmt.Sprintf("/v2/agents/%s/memory/search", agentID)
			} else {
				// Fallback to legacy endpoint
				endpoint = "/memory/search"
			}

			var response []api.Memory
			if err := client.GetWithQuery(endpoint, params, &response); err != nil {
				return fmt.Errorf("failed to search memories: %w", err)
			}

			if len(response) == 0 {
				utils.PrintInfo("No memories found matching your query")
				return nil
			}

			utils.PrintInfo("Found %d memories matching your query", len(response))

			var data []map[string]interface{}
			for _, memory := range response {
				data = append(data, map[string]interface{}{
					"ID":         memory.ID,
					"Session ID": memory.SessionID,
					"Summary":    memory.Summary,
					"Similarity": fmt.Sprintf("%.3f", memory.Similarity),
					"Created":    memory.CreatedAt.Format("2006-01-02 15:04:05"),
				})
			}

			return utils.PrintOutput(data, config.GetOutputFormat())
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "number of results to return")
	cmd.Flags().Float64VarP(&threshold, "threshold", "t", 0.7, "similarity threshold (0.0-1.0)")
	cmd.Flags().StringVar(&agentFlag, "agent", "", "agent name or ID")

	return cmd
}
