package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

// Server implements the MCP protocol over stdio using JSON-RPC 2.0.
type Server struct {
	initialized bool
}

// NewServer creates a new MCP server instance.
func NewServer() *Server {
	return &Server{}
}

// Run starts the MCP server, reading JSON-RPC messages from stdin and writing responses to stdout.
func (s *Server) Run() error {
	scanner := bufio.NewScanner(os.Stdin)
	// Increase buffer size for large messages
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.sendError(nil, -32700, "Parse error")
			continue
		}

		s.handleRequest(&req)
	}

	return scanner.Err()
}

func (s *Server) handleRequest(req *JSONRPCRequest) {
	switch req.Method {
	case "initialize":
		s.handleInitialize(req)
	case "notifications/initialized":
		// Client acknowledgment, no response needed
	case "tools/list":
		s.handleToolsList(req)
	case "tools/call":
		s.handleToolCall(req)
	case "ping":
		s.sendResult(req.ID, map[string]interface{}{})
	default:
		s.sendError(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
	}
}

func (s *Server) handleInitialize(req *JSONRPCRequest) {
	s.initialized = true
	s.sendResult(req.ID, InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{},
		},
		ServerInfo: ServerInfo{
			Name:    "aphelion",
			Version: "1.0.0",
		},
	})
}

func (s *Server) handleToolsList(req *JSONRPCRequest) {
	s.sendResult(req.ID, ToolsListResult{Tools: MCPTools})
}

func (s *Server) handleToolCall(req *JSONRPCRequest) {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "Invalid params")
		return
	}

	result := s.executeTool(params.Name, params.Arguments)
	s.sendResult(req.ID, result)
}

func (s *Server) executeTool(name string, args json.RawMessage) ToolCallResult {
	// Parse args into a map
	var argMap map[string]string
	if len(args) > 0 {
		_ = json.Unmarshal(args, &argMap)
	}
	if argMap == nil {
		argMap = make(map[string]string)
	}

	client := api.NewClient()

	switch name {
	case "aphelion_auth_status":
		return s.toolAuthStatus()
	case "aphelion_agent_init":
		return textResult("Agent init requires interactive mode. Run: aphelion agent init --name " + argMap["name"] + " --description '" + argMap["description"] + "'")
	case "aphelion_agents_list":
		return s.toolAgentsList(client)
	case "aphelion_agents_create":
		return s.toolAgentsCreate(client, argMap)
	case "aphelion_agents_inspect":
		return s.toolAgentsInspect(client, argMap)
	case "aphelion_deploy":
		return textResult("Deploy requires running from a project directory. Run: aphelion deploy")
	case "aphelion_invoke":
		return s.toolInvoke(client, argMap)
	case "aphelion_deployments_status":
		return s.toolDeploymentStatus(client, argMap)
	case "aphelion_deployments_logs":
		return s.toolDeploymentLogs(client, argMap)
	case "aphelion_memory_get":
		return s.toolMemoryGet(client, argMap)
	case "aphelion_memory_set":
		return s.toolMemorySet(client, argMap)
	case "aphelion_memory_search":
		return s.toolMemorySearch(client, argMap)
	case "aphelion_memory_list":
		return s.toolMemoryList(client, argMap)
	case "aphelion_tools_search":
		return s.toolToolsSearch(client, argMap)
	case "aphelion_tools_subscribe":
		return s.toolToolsSubscribe(client, argMap)
	case "aphelion_tools_describe":
		return s.toolToolsDescribe(client, argMap)
	case "aphelion_tools_try":
		return s.toolToolsTry(client, argMap)
	case "aphelion_env_set":
		return s.toolEnvSet(client, argMap)
	case "aphelion_env_list":
		return s.toolEnvList(client, argMap)
	case "aphelion_analytics":
		return s.toolAnalytics(client, argMap)
	case "aphelion_status":
		return s.toolStatus()
	default:
		return errorResult(fmt.Sprintf("Unknown tool: %s", name))
	}
}

// Tool handler implementations

func (s *Server) toolAuthStatus() ToolCallResult {
	if !config.IsAuthenticated() {
		return textResult("Not authenticated. Run: aphelion auth login")
	}
	email := config.GetUserEmail()
	accountID := config.GetAccountID()
	return textResult(fmt.Sprintf("Authenticated as %s (Account: %s)", email, accountID))
}

func (s *Server) toolAgentsList(client *api.Client) ToolCallResult {
	var resp json.RawMessage
	if err := client.Get("/v2/agents", &resp); err != nil {
		return errorResult(err.Error())
	}
	var pretty interface{}
	_ = json.Unmarshal(resp, &pretty)
	data, _ := json.MarshalIndent(pretty, "", "  ")
	return textResult(string(data))
}

func (s *Server) toolAgentsCreate(client *api.Client, args map[string]string) ToolCallResult {
	name := args["name"]
	desc := args["description"]
	if name == "" || desc == "" {
		return errorResult("name and description are required")
	}
	body := map[string]string{"name": name, "description": desc}
	var resp map[string]interface{}
	if err := client.Post("/v2/agents", body, &resp); err != nil {
		return errorResult(err.Error())
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return textResult(string(data))
}

func (s *Server) toolAgentsInspect(client *api.Client, args map[string]string) ToolCallResult {
	agent := args["agent"]
	if agent == "" {
		return errorResult("agent is required")
	}
	var resp map[string]interface{}
	if err := client.Get(fmt.Sprintf("/v2/agents/%s/inspect", agent), &resp); err != nil {
		return errorResult(err.Error())
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return textResult(string(data))
}

func (s *Server) toolInvoke(client *api.Client, args map[string]string) ToolCallResult {
	agent := args["agent"]
	input := args["input"]
	if agent == "" || input == "" {
		return errorResult("agent and input are required")
	}
	var inputObj interface{}
	if err := json.Unmarshal([]byte(input), &inputObj); err != nil {
		return errorResult(fmt.Sprintf("Invalid JSON input: %v", err))
	}
	var resp map[string]interface{}
	endpoint := fmt.Sprintf("/v2/agents/%s/invoke", agent)
	if err := client.Post(endpoint, inputObj, &resp); err != nil {
		return errorResult(err.Error())
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return textResult(string(data))
}

func (s *Server) toolDeploymentStatus(client *api.Client, args map[string]string) ToolCallResult {
	agent := args["agent"]
	if agent == "" {
		return errorResult("agent is required")
	}
	var resp map[string]interface{}
	if err := client.Get(fmt.Sprintf("/v2/agents/%s/deployment", agent), &resp); err != nil {
		return errorResult(err.Error())
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return textResult(string(data))
}

func (s *Server) toolDeploymentLogs(client *api.Client, args map[string]string) ToolCallResult {
	agent := args["agent"]
	if agent == "" {
		return errorResult("agent is required")
	}
	var resp map[string]interface{}
	if err := client.Get(fmt.Sprintf("/v2/agents/%s/logs", agent), &resp); err != nil {
		return errorResult(err.Error())
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return textResult(string(data))
}

func (s *Server) toolMemoryGet(client *api.Client, args map[string]string) ToolCallResult {
	agent := args["agent"]
	key := args["key"]
	if agent == "" || key == "" {
		return errorResult("agent and key are required")
	}
	var resp map[string]interface{}
	if err := client.Get(fmt.Sprintf("/v2/agents/%s/memory/%s", agent, key), &resp); err != nil {
		return errorResult(err.Error())
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return textResult(string(data))
}

func (s *Server) toolMemorySet(client *api.Client, args map[string]string) ToolCallResult {
	agent := args["agent"]
	key := args["key"]
	value := args["value"]
	if agent == "" || key == "" || value == "" {
		return errorResult("agent, key, and value are required")
	}
	var valueObj interface{}
	if err := json.Unmarshal([]byte(value), &valueObj); err != nil {
		return errorResult(fmt.Sprintf("Invalid JSON value: %v", err))
	}
	body := map[string]interface{}{"value": valueObj}
	if ttl := args["ttl"]; ttl != "" {
		body["ttl"] = ttl
	}
	if err := client.Put(fmt.Sprintf("/v2/agents/%s/memory/%s", agent, key), body, nil); err != nil {
		return errorResult(err.Error())
	}
	return textResult(fmt.Sprintf("Set memory key '%s' for agent '%s'", key, agent))
}

func (s *Server) toolMemorySearch(client *api.Client, args map[string]string) ToolCallResult {
	agent := args["agent"]
	query := args["query"]
	if agent == "" || query == "" {
		return errorResult("agent and query are required")
	}
	body := map[string]string{"query": query}
	var resp map[string]interface{}
	if err := client.Post(fmt.Sprintf("/v2/agents/%s/memory/search", agent), body, &resp); err != nil {
		return errorResult(err.Error())
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return textResult(string(data))
}

func (s *Server) toolMemoryList(client *api.Client, args map[string]string) ToolCallResult {
	agent := args["agent"]
	if agent == "" {
		return errorResult("agent is required")
	}
	var resp map[string]interface{}
	if err := client.Get(fmt.Sprintf("/v2/agents/%s/memory", agent), &resp); err != nil {
		return errorResult(err.Error())
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return textResult(string(data))
}

func (s *Server) toolToolsSearch(client *api.Client, args map[string]string) ToolCallResult {
	query := args["query"]
	if query == "" {
		return errorResult("query is required")
	}
	var resp map[string]interface{}
	params := map[string]string{"q": query}
	if err := client.GetWithQuery("/v2/tools/search", params, &resp); err != nil {
		return errorResult(err.Error())
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return textResult(string(data))
}

func (s *Server) toolToolsSubscribe(client *api.Client, args map[string]string) ToolCallResult {
	agent := args["agent"]
	tool := args["tool"]
	if agent == "" || tool == "" {
		return errorResult("agent and tool are required")
	}
	body := map[string]string{"tool": tool}
	if err := client.Post(fmt.Sprintf("/v2/agents/%s/tools/subscribe", agent), body, nil); err != nil {
		return errorResult(err.Error())
	}
	return textResult(fmt.Sprintf("Subscribed agent '%s' to tool '%s'", agent, tool))
}

func (s *Server) toolToolsDescribe(client *api.Client, args map[string]string) ToolCallResult {
	tool := args["tool"]
	if tool == "" {
		return errorResult("tool is required")
	}
	var resp map[string]interface{}
	if err := client.Get(fmt.Sprintf("/v2/tools/%s", tool), &resp); err != nil {
		return errorResult(err.Error())
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return textResult(string(data))
}

func (s *Server) toolToolsTry(client *api.Client, args map[string]string) ToolCallResult {
	tool := args["tool"]
	params := args["params"]
	if tool == "" || params == "" {
		return errorResult("tool and params are required")
	}
	var paramsObj interface{}
	if err := json.Unmarshal([]byte(params), &paramsObj); err != nil {
		return errorResult(fmt.Sprintf("Invalid JSON params: %v", err))
	}
	var resp map[string]interface{}
	if err := client.Post(fmt.Sprintf("/v2/tools/%s/execute", tool), paramsObj, &resp); err != nil {
		return errorResult(err.Error())
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return textResult(string(data))
}

func (s *Server) toolEnvSet(client *api.Client, args map[string]string) ToolCallResult {
	agent := args["agent"]
	key := args["key"]
	value := args["value"]
	if agent == "" || key == "" || value == "" {
		return errorResult("agent, key, and value are required")
	}
	body := map[string]string{"value": value}
	if err := client.Put(fmt.Sprintf("/v2/agents/%s/env/%s", agent, key), body, nil); err != nil {
		return errorResult(err.Error())
	}
	return textResult(fmt.Sprintf("Set %s for agent '%s'", key, agent))
}

func (s *Server) toolEnvList(client *api.Client, args map[string]string) ToolCallResult {
	agent := args["agent"]
	if agent == "" {
		return errorResult("agent is required")
	}
	var resp map[string]interface{}
	if err := client.Get(fmt.Sprintf("/v2/agents/%s/env", agent), &resp); err != nil {
		return errorResult(err.Error())
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return textResult(string(data))
}

func (s *Server) toolAnalytics(client *api.Client, args map[string]string) ToolCallResult {
	var resp map[string]interface{}
	endpoint := "/v2/analytics"
	if agent := args["agent"]; agent != "" {
		endpoint = fmt.Sprintf("/v2/analytics?agent=%s", agent)
	}
	if err := client.Get(endpoint, &resp); err != nil {
		return errorResult(err.Error())
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return textResult(string(data))
}

func (s *Server) toolStatus() ToolCallResult {
	if !config.IsProjectDir() {
		return errorResult("Not in an agent project directory. Run from a directory with .aphelion/config.yaml")
	}
	projCfg, err := config.LoadProjectConfig()
	if err != nil || projCfg == nil {
		return errorResult("Failed to load project config")
	}
	data, _ := json.MarshalIndent(projCfg, "", "  ")
	return textResult(string(data))
}

// Helpers

func textResult(text string) ToolCallResult {
	return ToolCallResult{
		Content: []ContentBlock{{Type: "text", Text: text}},
	}
}

func errorResult(text string) ToolCallResult {
	return ToolCallResult{
		Content: []ContentBlock{{Type: "text", Text: text}},
		IsError: true,
	}
}

func (s *Server) sendResult(id interface{}, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	data, _ := json.Marshal(resp)
	fmt.Fprintf(os.Stdout, "%s\n", data)
}

func (s *Server) sendError(id interface{}, code int, message string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &RPCError{Code: code, Message: message},
	}
	data, _ := json.Marshal(resp)
	fmt.Fprintf(os.Stdout, "%s\n", data)
}
