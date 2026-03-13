package memory

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
)

func newListCmd() *cobra.Command {
	var (
		limit     int
		sort      string
		dateFrom  string
		dateTo    string
		agentFlag string
		search    string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List memory entries",
		Long:  "List all memory entries with pagination and sorting options",
		Example: `  # List recent memories (from project directory)
  aphelion memory list

  # List memories for a specific agent
  aphelion memory list --agent review-management-agent

  # List with custom limit and sorting
  aphelion memory list --limit 10 --sort oldest

  # Search within memory list
  aphelion memory list --search "patient"

  # List memories from specific date range
  aphelion memory list --date-from 2023-01-01 --date-to 2023-12-31`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsAuthenticated() {
				return fmt.Errorf("session expired. Run: aphelion auth login")
			}

			client := api.NewClient()

			// Build query params
			params := map[string]string{
				"limit": strconv.Itoa(limit),
				"sort":  sort,
			}
			if dateFrom != "" {
				params["date_from"] = dateFrom
			}
			if dateTo != "" {
				params["date_to"] = dateTo
			}
			if search != "" {
				params["search"] = search
			}

			// Determine endpoint based on agent context
			var endpoint string
			agentID, agentErr := resolveAgentID(agentFlag)
			if agentErr == nil && agentID != "" {
				// Agent-scoped memory
				endpoint = fmt.Sprintf("/v2/agents/%s/memory", agentID)
			} else {
				// Fallback to legacy endpoint
				if limit != 10 || sort != "newest" || dateFrom != "" || dateTo != "" || search != "" {
					endpoint = "/memory/paginated"
				} else {
					endpoint = "/memory"
				}
			}

			var memories []api.Memory

			var response api.MemoriesResponse
			if err := client.GetWithQuery(endpoint, params, &response); err != nil {
				// Fallback to basic endpoint for legacy
				if agentID == "" {
					if err := client.Get("/memory", &response); err != nil {
						return fmt.Errorf("failed to list memories: %w", err)
					}
				} else {
					return fmt.Errorf("failed to list memories: %w", err)
				}
			}
			memories = response.Memories

			if len(memories) == 0 {
				utils.PrintInfo("No memories found")
				return nil
			}

			utils.PrintInfo("Found %d memories", len(memories))

			var data []map[string]interface{}
			for _, memory := range memories {
				data = append(data, map[string]interface{}{
					"ID":         memory.ID,
					"Session ID": memory.SessionID,
					"Summary":    memory.Summary,
					"Created":    memory.CreatedAt.Format("2006-01-02 15:04:05"),
				})
			}

			return utils.PrintOutput(data, config.GetOutputFormat())
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "number of memories to return (1-100)")
	cmd.Flags().StringVarP(&sort, "sort", "s", "newest", "sort order (newest, oldest)")
	cmd.Flags().StringVar(&dateFrom, "date-from", "", "filter from date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&dateTo, "date-to", "", "filter to date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&agentFlag, "agent", "", "agent name or ID")
	cmd.Flags().StringVar(&search, "search", "", "search term to filter memories")

	return cmd
}
