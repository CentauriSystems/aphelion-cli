package mcp

// MCPTools defines all tools exposed by the Aphelion MCP server.
var MCPTools = []ToolDef{
	{
		Name:        "aphelion_auth_status",
		Description: "Check if the user is authenticated with Aphelion",
		InputSchema: InputSchema{Type: "object"},
	},
	{
		Name:        "aphelion_agent_init",
		Description: "Initialize a new Aphelion agent project",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"name":        {Type: "string", Description: "Agent name in kebab-case"},
				"description": {Type: "string", Description: "Agent description"},
				"language":    {Type: "string", Description: "Language: python, node, or go"},
			},
			Required: []string{"name", "description"},
		},
	},
	{
		Name:        "aphelion_agents_list",
		Description: "List all agent identities on the account",
		InputSchema: InputSchema{Type: "object"},
	},
	{
		Name:        "aphelion_agents_create",
		Description: "Create a new agent identity",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"name":        {Type: "string", Description: "Agent name"},
				"description": {Type: "string", Description: "Agent description"},
			},
			Required: []string{"name", "description"},
		},
	},
	{
		Name:        "aphelion_agents_inspect",
		Description: "Get detailed information about an agent",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"agent": {Type: "string", Description: "Agent name or ID"},
			},
			Required: []string{"agent"},
		},
	},
	{
		Name:        "aphelion_deploy",
		Description: "Deploy the current agent project to Aphelion Cloud",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"region": {Type: "string", Description: "Deployment region (default: us-central1)"},
				"public": {Type: "string", Description: "Set to 'true' to list in marketplace"},
			},
		},
	},
	{
		Name:        "aphelion_invoke",
		Description: "Invoke a deployed agent with input",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"agent": {Type: "string", Description: "Agent name"},
				"input": {Type: "string", Description: "JSON input string"},
			},
			Required: []string{"agent", "input"},
		},
	},
	{
		Name:        "aphelion_deployments_status",
		Description: "Get deployment status for an agent",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"agent": {Type: "string", Description: "Agent name or ID"},
			},
			Required: []string{"agent"},
		},
	},
	{
		Name:        "aphelion_deployments_logs",
		Description: "Get recent logs from a deployed agent",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"agent": {Type: "string", Description: "Agent name"},
				"limit": {Type: "string", Description: "Number of log lines (default: 100)"},
			},
			Required: []string{"agent"},
		},
	},
	{
		Name:        "aphelion_memory_get",
		Description: "Read a memory key from an agent's memory",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"agent": {Type: "string", Description: "Agent name or ID"},
				"key":   {Type: "string", Description: "Memory key"},
			},
			Required: []string{"agent", "key"},
		},
	},
	{
		Name:        "aphelion_memory_set",
		Description: "Write a memory key to an agent's memory",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"agent": {Type: "string", Description: "Agent name or ID"},
				"key":   {Type: "string", Description: "Memory key"},
				"value": {Type: "string", Description: "JSON value"},
				"ttl":   {Type: "string", Description: "Time to live (e.g. 7d, 24h)"},
			},
			Required: []string{"agent", "key", "value"},
		},
	},
	{
		Name:        "aphelion_memory_search",
		Description: "Semantic search across an agent's memory",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"agent": {Type: "string", Description: "Agent name or ID"},
				"query": {Type: "string", Description: "Search query"},
			},
			Required: []string{"agent", "query"},
		},
	},
	{
		Name:        "aphelion_memory_list",
		Description: "List memory entries for an agent",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"agent": {Type: "string", Description: "Agent name or ID"},
				"limit": {Type: "string", Description: "Max entries to return"},
			},
			Required: []string{"agent"},
		},
	},
	{
		Name:        "aphelion_tools_search",
		Description: "Search the Aphelion tool marketplace",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"query": {Type: "string", Description: "Search query"},
			},
			Required: []string{"query"},
		},
	},
	{
		Name:        "aphelion_tools_subscribe",
		Description: "Subscribe an agent to a tool",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"agent": {Type: "string", Description: "Agent name or ID"},
				"tool":  {Type: "string", Description: "Tool name"},
			},
			Required: []string{"agent", "tool"},
		},
	},
	{
		Name:        "aphelion_tools_describe",
		Description: "Get detailed description and schema for a tool",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"tool": {Type: "string", Description: "Tool name (e.g. twilio.messaging.create_message)"},
			},
			Required: []string{"tool"},
		},
	},
	{
		Name:        "aphelion_tools_try",
		Description: "Execute a tool directly for testing",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"tool":   {Type: "string", Description: "Tool name"},
				"params": {Type: "string", Description: "JSON params string"},
			},
			Required: []string{"tool", "params"},
		},
	},
	{
		Name:        "aphelion_env_set",
		Description: "Set an environment variable for a deployed agent",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"agent": {Type: "string", Description: "Agent name or ID"},
				"key":   {Type: "string", Description: "Variable name"},
				"value": {Type: "string", Description: "Variable value"},
			},
			Required: []string{"agent", "key", "value"},
		},
	},
	{
		Name:        "aphelion_env_list",
		Description: "List environment variable keys for a deployed agent",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"agent": {Type: "string", Description: "Agent name or ID"},
			},
			Required: []string{"agent"},
		},
	},
	{
		Name:        "aphelion_analytics",
		Description: "Get usage analytics for the account or a specific agent",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"agent": {Type: "string", Description: "Agent name (optional, for agent-specific analytics)"},
			},
		},
	},
	{
		Name:        "aphelion_status",
		Description: "Get current project status including deployment, tools, memory, and execution stats",
		InputSchema: InputSchema{Type: "object"},
	},
}
