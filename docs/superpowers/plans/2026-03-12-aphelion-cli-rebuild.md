# Aphelion CLI Rebuild — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the Aphelion CLI from its current partial state into the full SDK described in CLAUDE.md — covering auth, agent lifecycle, deployment, memory, tools, MCP server, and all supporting commands.

**Architecture:** Cobra CLI with `cmd/` per command group, shared `pkg/api` HTTP client, `pkg/config` for config management, `pkg/auth` for Auth0 PKCE + token refresh. New command groups: `agents`, `deploy`, `deployments`, `invoke`, `env`, `schedule`, `status`, `mcp`. Python SDK scaffolded at `agent init` time and published as `aphelion-sdk`.

**Tech Stack:** Go 1.21, Cobra, Viper, Auth0 PKCE, MCP (stdio JSON-RPC), Python SDK

---

## Codebase Inventory — What Exists Today

```
Existing files (31 Go files):
  cmd/root.go                    — Root command, registers subcommands
  cmd/version.go                 — Version printing
  cmd/completion.go              — Shell completions
  cmd/agent/agent.go             — Agent parent command
  cmd/agent/init.go              — Basic scaffold (no interactivity, hardcoded template)
  cmd/agent/run.go               — Run agent locally (cron, daemon work; missing --input)
  cmd/auth/auth.go               — Auth parent command
  cmd/auth/login.go              — PKCE + legacy login (works, missing refresh)
  cmd/auth/logout.go             — Clear tokens
  cmd/auth/profile.go            — Show profile
  cmd/auth/register.go           — Legacy register (REMOVE)
  cmd/auth/oauth.go              — Get OAuth config
  cmd/analytics/analytics.go     — Analytics parent
  cmd/analytics/sessions.go      — Session analytics
  cmd/analytics/tools.go         — Tool analytics
  cmd/analytics/user.go          — User analytics
  cmd/memory/memory.go           — Memory parent
  cmd/memory/list.go             — List memories
  cmd/memory/search.go           — Search memories
  cmd/memory/clear.go            — Clear all memories
  cmd/memory/stats.go            — Memory stats
  cmd/registry/registry.go       — Registry parent
  cmd/registry/add_openapi.go    — Upload OpenAPI spec
  cmd/registry/create.go         — Create service
  cmd/registry/delete.go         — Delete service
  cmd/registry/get.go            — Get service details
  cmd/registry/list.go           — List services
  cmd/registry/my_services.go    — List own services
  cmd/tools/tools.go             — Tools parent
  cmd/tools/describe.go          — Describe a tool
  cmd/tools/try.go               — Try executing a tool
  pkg/api/client.go              — HTTP client (GET, POST, DELETE, GetWithQuery)
  pkg/api/types.go               — Request/response types
  pkg/auth/oauth.go              — PKCE flow + callback server
  pkg/auth/token.go              — Token exchange + user info
  pkg/config/config.go           — Global config (no project config, no refresh token)
  internal/utils/browser.go      — Open browser
  internal/utils/output.go       — Print helpers
  internal/utils/spinner.go      — Spinner
```

**Critical bugs in existing code:**
- API URL hardcoded to `https://api.aphelion.exmplr.ai` in 3 places (must be `https://api.aphl.ai`)
- Success redirect URL in `pkg/auth/oauth.go:155` points to `aphelion.exmplr.ai`
- No refresh token storage or auto-refresh
- Config struct missing `RefreshToken`, `ExpiresAt`, `AccountID`
- `TokenResponse` missing `RefreshToken` field
- `agent init` creates a non-functional template (no SDK, hardcoded "Multiple Sclerosis")
- API client has no `PUT` or `PATCH` methods

---

## Phase Breakdown

This plan is split into 6 phases, each producing working, testable software. Phases map to the CLAUDE.md implementation order.

| Phase | Scope | Tasks | New Files |
|-------|-------|-------|-----------|
| 1 | Foundation: API URL fix, config expansion, auth update | 1–5 | ~3 |
| 2 | Agent identity: `agents` commands + permissions | 6–9 | ~12 |
| 3 | Agent dev: `agent init` rewrite, `agent run` update, memory | 10–14 | ~8 |
| 4 | Deployment: `deploy`, `deployments`, `invoke`, `env`, `schedule` | 15–22 | ~14 |
| 5 | Discovery: `tools` updates, analytics, `status`, utilities | 23–28 | ~10 |
| 6 | MCP server + README | 29–31 | ~4 |

---

## Chunk 1: Foundation (Tasks 1–5)

### Task 1: Fix API URL Everywhere

**Files:**
- Modify: `pkg/config/config.go:26`
- Modify: `cmd/root.go:71`
- Modify: `cmd/agent/init.go:38`
- Modify: `pkg/auth/oauth.go:155`

- [ ] **Step 1: Fix `pkg/config/config.go`**

Change line 26:
```go
viper.SetDefault("api_url", "https://api.aphl.ai")
```

- [ ] **Step 2: Fix `cmd/root.go`**

Change line 71:
```go
rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "https://api.aphl.ai", "API base URL")
```

- [ ] **Step 3: Fix `cmd/agent/init.go`**

Change the config template line 38:
```go
  api_url: "https://api.aphl.ai"
```

- [ ] **Step 4: Fix success redirect in `pkg/auth/oauth.go`**

Change line 155:
```go
w.Header().Set("Location", "https://console.aphl.ai/auth/success")
```

- [ ] **Step 5: Build and verify**

Run: `go build -o aphelion .`
Expected: builds without errors

- [ ] **Step 6: Commit**

```bash
git add pkg/config/config.go cmd/root.go cmd/agent/init.go pkg/auth/oauth.go
git commit -m "fix: update API URL from api.aphelion.exmplr.ai to api.aphl.ai"
```

---

### Task 2: Expand Config Struct for Auth

**Files:**
- Modify: `pkg/config/config.go`
- Modify: `pkg/auth/token.go`

- [ ] **Step 1: Update Config struct to add RefreshToken, ExpiresAt, AccountID**

In `pkg/config/config.go`, replace the Config struct:
```go
type Config struct {
	APIUrl       string    `yaml:"api_url" mapstructure:"api_url"`
	Output       string    `yaml:"output" mapstructure:"output"`
	Auth         AuthConfig `yaml:"auth" mapstructure:"auth"`
	// Legacy fields for backward compat reads
	AccessToken  string    `yaml:"access_token,omitempty" mapstructure:"access_token"`
	UserID       string    `yaml:"user_id,omitempty" mapstructure:"user_id"`
	Email        string    `yaml:"email,omitempty" mapstructure:"email"`
	Username     string    `yaml:"username,omitempty" mapstructure:"username"`
	LastLogin    time.Time `yaml:"last_login,omitempty" mapstructure:"last_login"`
}

type AuthConfig struct {
	AccessToken  string `yaml:"access_token" mapstructure:"access_token"`
	RefreshToken string `yaml:"refresh_token" mapstructure:"refresh_token"`
	ExpiresAt    string `yaml:"expires_at" mapstructure:"expires_at"`
	UserEmail    string `yaml:"user_email" mapstructure:"user_email"`
	AccountID    string `yaml:"account_id" mapstructure:"account_id"`
	Username     string `yaml:"username" mapstructure:"username"`
}
```

- [ ] **Step 2: Update InitConfig to migrate legacy fields into Auth struct**

If `access_token` is set at top level but `auth.access_token` is not, copy it over. This handles existing configs gracefully.

- [ ] **Step 3: Update SetAuth to write RefreshToken and ExpiresAt**

```go
func SetAuth(token, refreshToken, userID, email, username string, expiresIn int) error {
	config := GetConfig()
	config.Auth.AccessToken = token
	config.Auth.RefreshToken = refreshToken
	config.Auth.UserEmail = email
	config.Auth.AccountID = userID
	config.Auth.Username = username
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
	config.Auth.ExpiresAt = expiresAt.Format(time.RFC3339)
	return SaveConfig()
}
```

- [ ] **Step 4: Update GetAccessToken to check expiry and auto-refresh**

```go
func GetAccessToken() string {
	config := GetConfig()
	// Check if token exists
	token := config.Auth.AccessToken
	if token == "" {
		token = config.AccessToken // legacy fallback
	}
	if token == "" {
		return ""
	}
	// Check expiry
	if config.Auth.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, config.Auth.ExpiresAt)
		if err == nil && time.Now().After(expiresAt) {
			// Token expired — try refresh
			if config.Auth.RefreshToken != "" {
				if err := RefreshAccessToken(); err != nil {
					return "" // refresh failed, user must re-login
				}
				return GetConfig().Auth.AccessToken
			}
			return "" // no refresh token
		}
	}
	return token
}
```

- [ ] **Step 5: Add RefreshAccessToken function**

In `pkg/config/config.go` or a new `pkg/auth/refresh.go`:
```go
func RefreshAccessToken() error {
	// POST to Auth0 token endpoint with grant_type=refresh_token
	// Update config with new access_token and expires_at
	// Save config
}
```

- [ ] **Step 6: Update TokenResponse in `pkg/auth/token.go`**

Add `RefreshToken` field:
```go
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}
```

- [ ] **Step 7: Update all callers of SetAuth (login.go)**

Update `cmd/auth/login.go:completeOAuthFlow` to pass refresh token and expires_in.

- [ ] **Step 8: Update ClearAuth to clear new fields**

- [ ] **Step 9: Build and verify**

Run: `go build -o aphelion .`

- [ ] **Step 10: Commit**

```bash
git commit -m "feat: expand config for refresh tokens and auto-refresh"
```

---

### Task 3: Add `auth status` and `auth token` Commands

**Files:**
- Create: `cmd/auth/status.go`
- Create: `cmd/auth/token.go`
- Modify: `cmd/auth/auth.go` (register new subcommands)

- [ ] **Step 1: Create `cmd/auth/status.go`**

Shows: auth context (human/agent/api-key), token expiry, user email.

```go
func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.GetConfig()
			if !config.IsAuthenticated() {
				fmt.Println("Not authenticated. Run: aphelion auth login")
				return nil
			}
			fmt.Printf("Authenticated as: %s\n", cfg.Auth.UserEmail)
			fmt.Printf("Account ID:       %s\n", cfg.Auth.AccountID)
			if cfg.Auth.ExpiresAt != "" {
				expiresAt, _ := time.Parse(time.RFC3339, cfg.Auth.ExpiresAt)
				if time.Now().Before(expiresAt) {
					fmt.Printf("Token expires:    %s (valid)\n", expiresAt.Format(time.RFC822))
				} else {
					fmt.Printf("Token expires:    %s (EXPIRED)\n", expiresAt.Format(time.RFC822))
				}
			}
			return nil
		},
	}
}
```

- [ ] **Step 2: Create `cmd/auth/token.go`**

Prints bearer token for scripting/debugging.

```go
func newTokenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "token",
		Short: "Print current bearer token",
		RunE: func(cmd *cobra.Command, args []string) error {
			token := config.GetAccessToken()
			if token == "" {
				return fmt.Errorf("not authenticated. Run: aphelion auth login")
			}
			fmt.Print(token)
			return nil
		},
	}
}
```

- [ ] **Step 3: Register in `cmd/auth/auth.go`**

- [ ] **Step 4: Remove `cmd/auth/register.go` and `cmd/auth/oauth.go`**

Remove the `register` and `oauth` subcommands (per spec: "Remove `aphelion auth register` and `aphelion auth oauth`").

- [ ] **Step 5: Build and verify**

- [ ] **Step 6: Commit**

```bash
git commit -m "feat: add auth status/token commands, remove register/oauth"
```

---

### Task 4: Add PUT/PATCH to API Client

**Files:**
- Modify: `pkg/api/client.go`

- [ ] **Step 1: Add Put method**

```go
func (c *Client) Put(endpoint string, body interface{}, result interface{}) error {
	resp, err := c.request("PUT", endpoint, body, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	return nil
}
```

- [ ] **Step 2: Add Patch method**

Same pattern as Put but with "PATCH".

- [ ] **Step 3: Add PostMultipart method (needed for deploy)**

For uploading tarballs:
```go
func (c *Client) PostMultipart(endpoint string, fieldName string, fileName string, fileData io.Reader, result interface{}) error
```

- [ ] **Step 4: Build and verify**

- [ ] **Step 5: Commit**

```bash
git commit -m "feat: add PUT, PATCH, and multipart POST to API client"
```

---

### Task 5: Add Project Config Support

**Files:**
- Create: `pkg/config/project.go`

- [ ] **Step 1: Create ProjectConfig struct**

```go
type ProjectConfig struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Version     string         `yaml:"version"`
	Language    string         `yaml:"language"`
	Agent       AgentConfig    `yaml:"agent"`
	Gateway     GatewayConfig  `yaml:"gateway"`
	Tools       ToolsConfig    `yaml:"tools"`
	Execution   ExecutionConfig `yaml:"execution"`
	Deployment  DeployConfig   `yaml:"deployment"`
	Schedule    ScheduleConfig `yaml:"schedule"`
	Logging     LoggingConfig  `yaml:"logging"`
}

type AgentConfig struct {
	ID           string `yaml:"id"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}
// ... (GatewayConfig, ToolsConfig, etc.)
```

- [ ] **Step 2: Add LoadProjectConfig and SaveProjectConfig functions**

Reads `.aphelion/config.yaml` in current directory. Returns nil if not found (not an error — means we're not in a project dir).

- [ ] **Step 3: Add IsProjectDir helper**

```go
func IsProjectDir() bool {
	_, err := os.Stat(".aphelion/config.yaml")
	return err == nil
}
```

- [ ] **Step 4: Build and verify**

- [ ] **Step 5: Commit**

```bash
git commit -m "feat: add project-level config support"
```

---

## Chunk 2: Agent Identity Management (Tasks 6–9)

### Task 6: Create `agents` Command Group

**Files:**
- Create: `cmd/agents/agents.go`
- Create: `cmd/agents/create.go`
- Create: `cmd/agents/list.go`
- Create: `cmd/agents/get.go`
- Modify: `cmd/root.go` (register `agents` command)
- Modify: `pkg/api/types.go` (add agent types)

- [ ] **Step 1: Add agent identity types to `pkg/api/types.go`**

```go
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
```

- [ ] **Step 2: Create `cmd/agents/agents.go`**

Parent command that registers create, list, get, update, rotate-secret, suspend, activate, delete, inspect as subcommands.

- [ ] **Step 3: Create `cmd/agents/create.go`**

`aphelion agents create --name <name> --description <desc>`
- POST `/v2/agents`
- Print client_id and client_secret (shown ONCE)
- If in project dir, offer to save to `.aphelion/config.yaml`

- [ ] **Step 4: Create `cmd/agents/list.go`**

`aphelion agents list`
- GET `/v2/agents`
- Table output: Name | ID | Status | Created | Last Active

- [ ] **Step 5: Create `cmd/agents/get.go`**

`aphelion agents get <name-or-id>`
- GET `/v2/agents/{id}`
- Full detail display

- [ ] **Step 6: Register in `cmd/root.go`**

Add import and `rootCmd.AddCommand(agents.NewAgentsCmd())`

- [ ] **Step 7: Build and verify**

- [ ] **Step 8: Commit**

```bash
git commit -m "feat: add agents create/list/get commands"
```

---

### Task 7: Agent Lifecycle Commands

**Files:**
- Create: `cmd/agents/update.go`
- Create: `cmd/agents/rotate_secret.go`
- Create: `cmd/agents/suspend.go`
- Create: `cmd/agents/activate.go`
- Create: `cmd/agents/delete.go`
- Create: `cmd/agents/inspect.go`

- [ ] **Step 1: Create `cmd/agents/update.go`**

`aphelion agents update <name-or-id> --description <desc>`
- PATCH `/v2/agents/{id}`

- [ ] **Step 2: Create `cmd/agents/rotate_secret.go`**

Confirmation prompt → POST `/v2/agents/{id}/rotate-secret` → print new secret → update project config if applicable.

- [ ] **Step 3: Create `cmd/agents/suspend.go` and `cmd/agents/activate.go`**

POST `/v2/agents/{id}/suspend` and POST `/v2/agents/{id}/activate`

- [ ] **Step 4: Create `cmd/agents/delete.go`**

Requires typing agent name for confirmation.

- [ ] **Step 5: Create `cmd/agents/inspect.go`**

GET `/v2/agents/{id}/inspect` — shows subscriptions, memory count, permissions, deployment, recent executions.

- [ ] **Step 6: Build and verify**

- [ ] **Step 7: Commit**

```bash
git commit -m "feat: add agents update/rotate-secret/suspend/activate/delete/inspect"
```

---

### Task 8: Agent Permission Commands

**Files:**
- Create: `cmd/agents/grant.go`
- Create: `cmd/agents/revoke.go`
- Create: `cmd/agents/permissions.go`
- Create: `cmd/agents/grants.go`

- [ ] **Step 1: Create `cmd/agents/grant.go`**

`aphelion agents grant --from <grantee> --to <resource> --actions read[,write] [--expires DATE]`
- POST `/v2/agents/permissions/grant`

- [ ] **Step 2: Create `cmd/agents/revoke.go`**

`aphelion agents revoke --from <grantee> --to <resource>`
- DELETE `/v2/agents/permissions/grant`

- [ ] **Step 3: Create `cmd/agents/permissions.go`**

`aphelion agents permissions <name-or-id>`
- GET `/v2/agents/{id}/permissions` — who can access this agent's memory

- [ ] **Step 4: Create `cmd/agents/grants.go`**

`aphelion agents grants <name-or-id>`
- GET `/v2/agents/{id}/grants` — what this agent has access to

- [ ] **Step 5: Build and verify**

- [ ] **Step 6: Commit**

```bash
git commit -m "feat: add agent permission grant/revoke/list commands"
```

---

### Task 9: End-of-phase checkpoint

- [ ] **Step 1: Run full build**

```bash
go build -o aphelion . && ./aphelion --help
```

- [ ] **Step 2: Verify command tree**

```bash
./aphelion agents --help
./aphelion auth --help
```

- [ ] **Step 3: Commit any fixups**

---

## Chunk 3: Agent Development (Tasks 10–14)

### Task 10: Rewrite `agent init` — Interactive Scaffold

**Files:**
- Modify: `cmd/agent/init.go` (complete rewrite)
- Create: `cmd/agent/templates.go` (template strings for generated files)

This is the most important command. It must be interactive and generate a complete, working project.

- [ ] **Step 1: Create `cmd/agent/templates.go`**

Contains Go const strings for all generated file templates:
- `pythonAgentTemplate` — the full agent.py from CLAUDE.md spec
- `nodeAgentTemplate` — equivalent for Node.js
- `agentJSONTemplate` — agent.json manifest
- `envExampleTemplate` — .env.example
- `requirementsTemplate` — requirements.txt
- `packageJSONTemplate` — package.json (Node only)
- `readmeTemplate` — project README
- `testTemplate` — test scaffold
- `gitignoreTemplate` — .aphelion/.gitignore

- [ ] **Step 2: Rewrite `cmd/agent/init.go`**

Interactive flow:
1. Prompt for agent name (kebab-case validation)
2. Prompt for description
3. Prompt for language (python/node/go)
4. Prompt for tools to subscribe (calls `GET /v2/tools/search` as user types)
5. Ask "Create agent identity now? [Y/n]" → calls `POST /v2/agents`
6. Creates full directory structure per spec
7. Writes `.aphelion/config.yaml` with agent credentials
8. Writes `.aphelion/.gitignore`
9. Writes `agent.json`
10. Writes language-specific agent file
11. Writes `.env.example`
12. Writes `tests/` scaffold
13. Writes `README.md`
14. Prints next steps

- [ ] **Step 3: Build and test init flow**

```bash
go build -o aphelion . && ./aphelion agent init
```

- [ ] **Step 4: Commit**

```bash
git commit -m "feat: rewrite agent init with interactive scaffold and full templates"
```

---

### Task 11: Update `agent run` with --input

**Files:**
- Modify: `cmd/agent/run.go`

- [ ] **Step 1: Add --input and --input-file flags**

```go
cmd.Flags().StringVar(&input, "input", "", "JSON input to pass to agent")
cmd.Flags().StringVar(&inputFile, "input-file", "", "path to JSON input file")
```

- [ ] **Step 2: Add --watch flag for streaming logs**

- [ ] **Step 3: Before execution, exchange agent credentials for JWT**

Read `.aphelion/config.yaml`, extract `agent.client_id` and `agent.client_secret`, POST to `/auth/agent/token`, inject as `APHELION_API_TOKEN` env var into subprocess.

- [ ] **Step 4: Inject all APHELION_* env vars into subprocess**

```
APHELION_AGENT_ID, APHELION_SESSION_ID, APHELION_API_TOKEN, APHELION_API_URL
```

- [ ] **Step 5: After execution, print formatted output and duration**

- [ ] **Step 6: Build and verify**

- [ ] **Step 7: Commit**

```bash
git commit -m "feat: add --input/--input-file flags and JWT injection to agent run"
```

---

### Task 12: Update Memory Commands — get, set, delete

**Files:**
- Create: `cmd/memory/get.go`
- Create: `cmd/memory/set.go`
- Create: `cmd/memory/delete.go`
- Modify: `cmd/memory/memory.go` (register new subcommands)
- Modify: `cmd/memory/list.go` (add --agent flag)

- [ ] **Step 1: Create `cmd/memory/get.go`**

`aphelion memory get <key> [--agent <name>] [--from <other-agent>]`
- GET `/v2/agents/{agent_id}/memory/{key}`

- [ ] **Step 2: Create `cmd/memory/set.go`**

`aphelion memory set <key> <value> [--ttl 7d]`
- PUT `/v2/agents/{agent_id}/memory/{key}`

- [ ] **Step 3: Create `cmd/memory/delete.go`**

`aphelion memory delete <key>`
- DELETE `/v2/agents/{agent_id}/memory/{key}`

- [ ] **Step 4: Add --agent flag to all memory commands**

If in project dir, infer agent from `.aphelion/config.yaml`. Otherwise require `--agent`.

- [ ] **Step 5: Update list.go with --search and --limit flags**

- [ ] **Step 6: Build and verify**

- [ ] **Step 7: Commit**

```bash
git commit -m "feat: add memory get/set/delete and agent-scoped access"
```

---

### Task 13: Add Actionable Error Messages

**Files:**
- Create: `pkg/api/errors.go`
- Modify: `pkg/api/client.go`

- [ ] **Step 1: Create error mapping in `pkg/api/errors.go`**

Map HTTP status codes + endpoint patterns to actionable messages:
```go
func actionableError(statusCode int, endpoint string, body string) error {
	switch statusCode {
	case 401:
		return fmt.Errorf("session expired. Run: aphelion auth login")
	case 404:
		if strings.Contains(endpoint, "/agents/") {
			return fmt.Errorf("agent not found.\nList your agents: aphelion agents list")
		}
		// ... etc
	}
}
```

- [ ] **Step 2: Wire into client.go error handling**

Replace raw API error returns with actionable versions.

- [ ] **Step 3: Build and verify**

- [ ] **Step 4: Commit**

```bash
git commit -m "feat: actionable error messages for all API failures"
```

---

### Task 14: Checkpoint — verify Phase 1-3

- [ ] **Step 1: Build**

```bash
go build -o aphelion .
```

- [ ] **Step 2: Verify command tree**

```bash
./aphelion --help
./aphelion agent --help
./aphelion agents --help
./aphelion auth --help
./aphelion memory --help
```

- [ ] **Step 3: Test auth flow (manual)**

```bash
./aphelion auth login
./aphelion auth status
./aphelion whoami  # (not yet — added in phase 5)
```

- [ ] **Step 4: Commit any fixups**

---

## Chunk 4: Deployment Pipeline (Tasks 15–22)

### Task 15: `deploy` Command

**Files:**
- Create: `cmd/deploy/deploy.go`
- Modify: `cmd/root.go` (register)

- [ ] **Step 1: Create `cmd/deploy/deploy.go`**

`aphelion deploy [--agent name] [--public] [--dry-run] [--region us-central1]`

Steps:
1. Read `.aphelion/config.yaml` and `agent.json` — validate
2. Validate agent code syntax
3. Check tool subscriptions
4. Create tarball
5. Upload via POST `/v2/agents/{id}/deploy` (multipart)
6. Poll status with spinner
7. Health check
8. Update project config
9. Print endpoint + invocation example

- [ ] **Step 2: Add tarball creation helper**

```go
func createTarball(dir string) (io.Reader, int64, error)
```

- [ ] **Step 3: Register in root.go**

- [ ] **Step 4: Build and verify**

- [ ] **Step 5: Commit**

```bash
git commit -m "feat: add deploy command with full deployment pipeline"
```

---

### Task 16: `deployments` Command Group

**Files:**
- Create: `cmd/deployments/deployments.go`
- Create: `cmd/deployments/list.go`
- Create: `cmd/deployments/status.go`
- Create: `cmd/deployments/logs.go`
- Create: `cmd/deployments/history.go`
- Create: `cmd/deployments/rollback.go`
- Create: `cmd/deployments/stop.go`
- Create: `cmd/deployments/redeploy.go`
- Create: `cmd/deployments/delete.go`
- Modify: `cmd/root.go`

- [ ] **Step 1: Create parent `cmd/deployments/deployments.go`**

- [ ] **Step 2: Create list.go**

GET `/v2/deployments` → table: Agent | Status | Endpoint | Region | Last Deploy | Executions (24h)

- [ ] **Step 3: Create status.go**

GET `/v2/agents/{id}/deployment` — current project's deployment.

- [ ] **Step 4: Create logs.go**

GET `/v2/agents/{id}/logs[?follow=true]` — with `--follow` flag for streaming.

- [ ] **Step 5: Create history.go**

GET `/v2/agents/{id}/executions` — with `--limit` and `--status` filters.

- [ ] **Step 6: Create rollback.go, stop.go, redeploy.go, delete.go**

POST/DELETE to respective endpoints with confirmation prompts.

- [ ] **Step 7: Register in root.go**

- [ ] **Step 8: Build and verify**

- [ ] **Step 9: Commit**

```bash
git commit -m "feat: add deployments list/status/logs/history/rollback/stop/redeploy/delete"
```

---

### Task 17: `invoke` Command

**Files:**
- Create: `cmd/invoke/invoke.go`
- Modify: `cmd/root.go`

- [ ] **Step 1: Create `cmd/invoke/invoke.go`**

`aphelion invoke <agent-name> --input '{...}' [--input-file file] [--watch] [--output table]`
- POST `/v2/agents/{id}/invoke`
- Handles `owner/agent-name` format for marketplace agents
- `--watch` streams execution logs then prints result

- [ ] **Step 2: Register in root.go**

- [ ] **Step 3: Build and verify**

- [ ] **Step 4: Commit**

```bash
git commit -m "feat: add invoke command for calling deployed agents"
```

---

### Task 18: `env` Command Group

**Files:**
- Create: `cmd/env/env.go`
- Create: `cmd/env/set.go`
- Create: `cmd/env/list.go`
- Create: `cmd/env/delete.go`
- Create: `cmd/env/pull.go`
- Create: `cmd/env/push.go`
- Modify: `cmd/root.go`

- [ ] **Step 1: Create parent and subcommands**

- `set`: PUT `/v2/agents/{id}/env/{key}`
- `list`: GET `/v2/agents/{id}/env` (keys only, never values)
- `delete`: DELETE `/v2/agents/{id}/env/{key}`
- `pull`: GET `/v2/agents/{id}/env` → write to `.env` (masked values)
- `push`: Read `.env` → PUT `/v2/agents/{id}/env` (confirm before overwriting)

- [ ] **Step 2: Register in root.go**

- [ ] **Step 3: Build and verify**

- [ ] **Step 4: Commit**

```bash
git commit -m "feat: add env set/list/delete/pull/push for deployed agent secrets"
```

---

### Task 19: `schedule` Command Group

**Files:**
- Create: `cmd/schedule/schedule.go`
- Create: `cmd/schedule/set.go`
- Create: `cmd/schedule/get.go`
- Create: `cmd/schedule/enable.go`
- Create: `cmd/schedule/disable.go`
- Create: `cmd/schedule/remove.go`
- Modify: `cmd/root.go`

- [ ] **Step 1: Create parent and subcommands**

- `set`: POST `/v2/agents/{id}/schedule` with cron expression
- `get`: GET `/v2/agents/{id}/schedule` — shows cron, next 5 runs, enabled/disabled
- `enable`/`disable`: PATCH `/v2/agents/{id}/schedule`
- `remove`: DELETE `/v2/agents/{id}/schedule`

- [ ] **Step 2: Register in root.go**

- [ ] **Step 3: Build and verify**

- [ ] **Step 4: Commit**

```bash
git commit -m "feat: add schedule set/get/enable/disable/remove for deployed agents"
```

---

### Task 20: End-to-end test — deployment pipeline

- [ ] **Step 1: Build**

```bash
go build -o aphelion .
```

- [ ] **Step 2: Verify full command tree**

```bash
./aphelion deploy --help
./aphelion deployments --help
./aphelion invoke --help
./aphelion env --help
./aphelion schedule --help
```

- [ ] **Step 3: Test deploy dry-run (if API is available)**

```bash
./aphelion deploy --dry-run
```

- [ ] **Step 4: Commit any fixups**

---

## Chunk 5: Discovery & Utilities (Tasks 21–26)

### Task 21: Update `tools` Commands

**Files:**
- Create: `cmd/tools/search.go`
- Create: `cmd/tools/subscribe.go`
- Create: `cmd/tools/unsubscribe.go`
- Create: `cmd/tools/list.go`
- Create: `cmd/tools/marketplace.go`
- Modify: `cmd/tools/tools.go` (register)

- [ ] **Step 1: Create `cmd/tools/search.go`**

`aphelion tools search "sms messaging"`
- GET `/v2/tools/search?q=...`

- [ ] **Step 2: Create `cmd/tools/subscribe.go` and `cmd/tools/unsubscribe.go`**

- POST `/v2/agents/{id}/tools/subscribe` with tool name
- DELETE `/v2/agents/{id}/tools/{tool}`

- [ ] **Step 3: Create `cmd/tools/list.go`**

`aphelion tools list` — lists tools subscribed by current agent.
- GET `/v2/agents/{id}/tools`

- [ ] **Step 4: Create `cmd/tools/marketplace.go`**

`aphelion tools marketplace [--category X] [--free] [--paid]`
- GET `/v2/tools/marketplace`

- [ ] **Step 5: Update try.go to add --dry-run if not already working**

- [ ] **Step 6: Build and verify**

- [ ] **Step 7: Commit**

```bash
git commit -m "feat: add tools search/subscribe/unsubscribe/list/marketplace"
```

---

### Task 22: Update `registry` Commands

**Files:**
- Create: `cmd/registry/update.go`
- Create: `cmd/registry/publish.go`
- Create: `cmd/registry/unpublish.go`
- Create: `cmd/registry/earnings.go`
- Modify: `cmd/registry/registry.go`

- [ ] **Step 1: Create update, publish, unpublish, earnings subcommands**

Each maps to a simple API call (PATCH, POST, DELETE, GET).

- [ ] **Step 2: Register in registry.go**

- [ ] **Step 3: Build and verify**

- [ ] **Step 4: Commit**

```bash
git commit -m "feat: add registry update/publish/unpublish/earnings commands"
```

---

### Task 23: Update `analytics` Commands

**Files:**
- Modify: `cmd/analytics/analytics.go`
- Create: `cmd/analytics/executions.go`
- Create: `cmd/analytics/earnings.go`

- [ ] **Step 1: Add --agent, --last, --status flags to existing commands**

- [ ] **Step 2: Create executions.go**

`aphelion analytics executions [--agent X] [--status failed] [--last 7d]`

- [ ] **Step 3: Create earnings.go**

`aphelion analytics earnings [--last 30d]`

- [ ] **Step 4: Build and verify**

- [ ] **Step 5: Commit**

```bash
git commit -m "feat: add analytics executions/earnings and agent-scoped filters"
```

---

### Task 24: `status` Command

**Files:**
- Create: `cmd/status/status.go`
- Modify: `cmd/root.go`

- [ ] **Step 1: Create `cmd/status/status.go`**

Reads project config, queries multiple API endpoints, displays formatted dashboard per spec.

- [ ] **Step 2: Register in root.go**

- [ ] **Step 3: Build and verify**

- [ ] **Step 4: Commit**

```bash
git commit -m "feat: add status command showing project dashboard"
```

---

### Task 25: Utility Commands — `whoami`, `open`, `quickstart`

**Files:**
- Create: `cmd/whoami.go`
- Create: `cmd/open.go`
- Create: `cmd/quickstart.go`
- Modify: `cmd/root.go`

- [ ] **Step 1: Create `cmd/whoami.go`**

GET `/auth/test-profile` → print email, account ID, plan, agent count, API count.

- [ ] **Step 2: Create `cmd/open.go`**

`aphelion open [agents|marketplace|history|docs]`
Opens `console.aphl.ai` or subpath in browser.

- [ ] **Step 3: Create `cmd/quickstart.go`**

Interactive tutorial: auth → init → subscribe → run → deploy. Guided walkthrough.

- [ ] **Step 4: Register all in root.go**

- [ ] **Step 5: Build and verify**

- [ ] **Step 6: Commit**

```bash
git commit -m "feat: add whoami, open, and quickstart commands"
```

---

### Task 26: Checkpoint — verify Phase 5

- [ ] **Step 1: Full build and command tree check**

```bash
go build -o aphelion . && ./aphelion --help
```

- [ ] **Step 2: Verify all new commands show in help**

---

## Chunk 6: MCP Server + README (Tasks 27–29)

### Task 27: MCP Server — `mcp serve`

**Files:**
- Create: `cmd/mcp/mcp.go`
- Create: `cmd/mcp/serve.go`
- Create: `cmd/mcp/config.go`
- Create: `pkg/mcp/server.go`
- Create: `pkg/mcp/protocol.go`
- Create: `pkg/mcp/tools.go`
- Modify: `cmd/root.go`

This is the highest-leverage feature — enables Claude to use Aphelion as a tool provider.

- [ ] **Step 1: Create `pkg/mcp/protocol.go`**

JSON-RPC 2.0 message types for MCP:
```go
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
```

MCP lifecycle methods: `initialize`, `tools/list`, `tools/call`.

- [ ] **Step 2: Create `pkg/mcp/tools.go`**

Define all MCP tool definitions as a Go slice:
```go
var MCPTools = []MCPToolDef{
	{Name: "aphelion_auth_status", Description: "Check authentication status", InputSchema: ...},
	{Name: "aphelion_agent_init", Description: "Initialize new agent project", InputSchema: ...},
	// ... all 20+ tools from spec
}
```

Each tool maps to an existing CLI function. The tool handler calls the same Go code the CLI command uses.

- [ ] **Step 3: Create `pkg/mcp/server.go`**

Main server loop:
1. Read JSON-RPC from stdin line-by-line
2. Dispatch to handler based on method
3. Write JSON-RPC response to stdout
4. Handle `initialize` → return server info + capabilities
5. Handle `tools/list` → return MCPTools
6. Handle `tools/call` → dispatch to tool handler, return result

```go
func (s *Server) Run() error {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var req JSONRPCRequest
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			s.sendError(nil, -32700, "Parse error")
			continue
		}
		s.handleRequest(&req)
	}
	return scanner.Err()
}
```

- [ ] **Step 4: Create `cmd/mcp/serve.go`**

```go
func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start MCP server for AI assistant integration",
		RunE: func(cmd *cobra.Command, args []string) error {
			server := mcp.NewServer()
			return server.Run()
		},
	}
}
```

- [ ] **Step 5: Create `cmd/mcp/config.go`**

`aphelion mcp config` — prints Claude Desktop config JSON.

- [ ] **Step 6: Register in root.go**

- [ ] **Step 7: Build and test MCP server**

```bash
go build -o aphelion .
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"capabilities":{}}}' | ./aphelion mcp serve
```

Expected: JSON response with server capabilities.

- [ ] **Step 8: Commit**

```bash
git commit -m "feat: add MCP server for Claude/AI assistant integration"
```

---

### Task 28: Update README

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Rewrite README with new command reference**

Cover: installation, quickstart, all command groups, MCP setup, Python SDK usage.

- [ ] **Step 2: Commit**

```bash
git commit -m "docs: update README with full command reference and quickstart"
```

---

### Task 29: Final End-to-End Verification

- [ ] **Step 1: Full build**

```bash
go build -o aphelion .
```

- [ ] **Step 2: Verify all commands listed in CLAUDE.md exist in help output**

```bash
./aphelion --help
./aphelion auth --help
./aphelion agent --help
./aphelion agents --help
./aphelion memory --help
./aphelion tools --help
./aphelion registry --help
./aphelion analytics --help
./aphelion deploy --help
./aphelion deployments --help
./aphelion invoke --help
./aphelion env --help
./aphelion schedule --help
./aphelion status --help
./aphelion mcp --help
./aphelion whoami --help
./aphelion open --help
./aphelion quickstart --help
```

- [ ] **Step 3: Test MCP server initialization**

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./aphelion mcp serve
```

- [ ] **Step 4: Run through the E2E flow from CLAUDE.md spec (manual)**

Auth → init → subscribe tools → run → deploy → invoke → logs → status

- [ ] **Step 5: Final commit**

```bash
git commit -m "chore: final verification of all CLI commands"
```

---

## File Summary

**New files to create (~45):**
```
cmd/agents/agents.go
cmd/agents/create.go
cmd/agents/list.go
cmd/agents/get.go
cmd/agents/update.go
cmd/agents/rotate_secret.go
cmd/agents/suspend.go
cmd/agents/activate.go
cmd/agents/delete.go
cmd/agents/inspect.go
cmd/agents/grant.go
cmd/agents/revoke.go
cmd/agents/permissions.go
cmd/agents/grants.go
cmd/agent/templates.go
cmd/auth/status.go
cmd/auth/token.go
cmd/deploy/deploy.go
cmd/deployments/deployments.go
cmd/deployments/list.go
cmd/deployments/status.go
cmd/deployments/logs.go
cmd/deployments/history.go
cmd/deployments/rollback.go
cmd/deployments/stop.go
cmd/deployments/redeploy.go
cmd/deployments/delete.go
cmd/invoke/invoke.go
cmd/env/env.go
cmd/env/set.go
cmd/env/list.go
cmd/env/delete.go
cmd/env/pull.go
cmd/env/push.go
cmd/schedule/schedule.go
cmd/schedule/set.go
cmd/schedule/get.go
cmd/schedule/enable.go
cmd/schedule/disable.go
cmd/schedule/remove.go
cmd/status/status.go
cmd/tools/search.go
cmd/tools/subscribe.go
cmd/tools/unsubscribe.go
cmd/tools/list.go
cmd/tools/marketplace.go
cmd/registry/update.go
cmd/registry/publish.go
cmd/registry/unpublish.go
cmd/registry/earnings.go
cmd/analytics/executions.go
cmd/analytics/earnings.go
cmd/whoami.go
cmd/open.go
cmd/quickstart.go
cmd/mcp/mcp.go
cmd/mcp/serve.go
cmd/mcp/config.go
pkg/mcp/server.go
pkg/mcp/protocol.go
pkg/mcp/tools.go
pkg/api/errors.go
pkg/config/project.go
pkg/auth/refresh.go
```

**Files to modify (~12):**
```
pkg/config/config.go
pkg/api/client.go
pkg/api/types.go
pkg/auth/token.go
pkg/auth/oauth.go
cmd/root.go
cmd/agent/init.go
cmd/agent/run.go
cmd/auth/auth.go
cmd/memory/memory.go
cmd/memory/list.go
cmd/tools/tools.go
cmd/registry/registry.go
cmd/analytics/analytics.go
README.md
```

**Files to delete (~2):**
```
cmd/auth/register.go
cmd/auth/oauth.go
```
