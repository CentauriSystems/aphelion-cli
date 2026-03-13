package api

import "time"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type Service struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Spec        map[string]interface{} `json:"spec"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type ServicesResponse struct {
	Services []Service `json:"services"`
	Total    int       `json:"total"`
}

type Memory struct {
	ID          string                 `json:"id"`
	SessionID   string                 `json:"session_id"`
	Summary     string                 `json:"summary"`
	Content     map[string]interface{} `json:"content"`
	CreatedAt   time.Time              `json:"created_at"`
	Similarity  float64                `json:"similarity,omitempty"`
}

type MemoriesResponse struct {
	Memories []Memory `json:"memories"`
	Total    int      `json:"total"`
	Cursor   string   `json:"cursor,omitempty"`
}

type MemoryStats struct {
	TotalMemories    int     `json:"total_memories"`
	TotalSessions    int     `json:"total_sessions"`
	AveragePerDay    float64 `json:"average_per_day"`
	OldestMemory     string  `json:"oldest_memory"`
	MostRecentMemory string  `json:"most_recent_memory"`
}

type Analytics struct {
	RequestMetrics RequestMetrics `json:"request_metrics"`
	UserMetrics    UserMetrics    `json:"user_metrics"`
	ToolMetrics    ToolMetrics    `json:"tool_metrics"`
	SessionMetrics SessionMetrics `json:"session_metrics"`
}

type RequestMetrics struct {
	TotalRequests    int     `json:"total_requests"`
	SuccessfulCount  int     `json:"successful_count"`
	ErrorCount       int     `json:"error_count"`
	AverageTime      float64 `json:"average_time"`
	SuccessRate      float64 `json:"success_rate"`
}

type UserMetrics struct {
	UniqueUsers   int `json:"unique_users"`
	ActiveUsers   int `json:"active_users"`
	NewUsers      int `json:"new_users"`
	ReturningUsers int `json:"returning_users"`
}

type ToolMetrics struct {
	TotalExecutions int         `json:"total_executions"`
	UniqueTools     int         `json:"unique_tools"`
	PopularTools    []ToolUsage `json:"popular_tools"`
}

type ToolUsage struct {
	Tool        string  `json:"tool"`
	Count       int     `json:"count"`
	SuccessRate float64 `json:"success_rate"`
}

type SessionMetrics struct {
	TotalSessions     int     `json:"total_sessions"`
	ActiveSessions    int     `json:"active_sessions"`
	AverageActivities float64 `json:"average_activities"`
	AverageDuration   float64 `json:"average_duration"`
}

// Agent identity types

type AgentIdentity struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	ClientID     string    `json:"client_id,omitempty"`
	ClientSecret string    `json:"client_secret,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	LastActive   time.Time `json:"last_active,omitempty"`
}

type AgentsResponse struct {
	Agents []AgentIdentity `json:"agents"`
	Total  int             `json:"total"`
}

type CreateAgentRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RotateSecretResponse struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type AgentInspection struct {
	Agent       AgentIdentity   `json:"agent"`
	Tools       []string        `json:"tools"`
	MemoryCount int             `json:"memory_count"`
	Permissions []AgentPermission `json:"permissions"`
	Deployment  *DeploymentInfo `json:"deployment,omitempty"`
	RecentExecs []ExecutionInfo `json:"recent_executions,omitempty"`
}

type AgentPermission struct {
	GranteeAgent  string   `json:"grantee_agent"`
	ResourceAgent string   `json:"resource_agent"`
	Actions       []string `json:"actions"`
	ExpiresAt     string   `json:"expires_at,omitempty"`
}

type DeploymentInfo struct {
	Status       string `json:"status"`
	Endpoint     string `json:"endpoint"`
	Region       string `json:"region"`
	LastDeployed string `json:"last_deployed"`
}

type ExecutionInfo struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	StartedAt string `json:"started_at"`
	Duration  string `json:"duration"`
}

type GrantRequest struct {
	FromAgent string   `json:"from_agent"`
	ToAgent   string   `json:"to_agent"`
	Actions   []string `json:"actions"`
	ExpiresAt string   `json:"expires_at,omitempty"`
}

// Deployment types

type DeploymentSummary struct {
	AgentName    string `json:"agent_name"`
	Status       string `json:"status"`
	Endpoint     string `json:"endpoint"`
	Region       string `json:"region"`
	LastDeployed string `json:"last_deployed"`
	Executions24 int    `json:"executions_24h"`
}

type DeploymentsResponse struct {
	Deployments []DeploymentSummary `json:"deployments"`
	Total       int                 `json:"total"`
}

type DeploymentStatus struct {
	AgentID        string `json:"agent_id"`
	AgentName      string `json:"agent_name"`
	Status         string `json:"status"`
	Endpoint       string `json:"endpoint"`
	Region         string `json:"region"`
	LastDeployed   string `json:"last_deployed"`
	ExecutionCount int    `json:"execution_count"`
	Version        string `json:"version"`
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

type LogsResponse struct {
	Logs []LogEntry `json:"logs"`
}

type ExecutionRecord struct {
	ID            string `json:"id"`
	Timestamp     string `json:"timestamp"`
	InputSummary  string `json:"input_summary"`
	OutputSummary string `json:"output_summary"`
	Duration      string `json:"duration"`
	Status        string `json:"status"`
}

type ExecutionsResponse struct {
	Executions []ExecutionRecord `json:"executions"`
	Total      int               `json:"total"`
}

type RollbackResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Message string `json:"message"`
}

type RedeployResponse struct {
	Status   string `json:"status"`
	Endpoint string `json:"endpoint"`
	Message  string `json:"message"`
}