# Aphelion CLI — Claude Implementation Instructions

You are updating the Aphelion CLI, a Go command-line tool at `github.com/Exmplr-AI/aphelion-cli`.

Read this entire file before touching any code. Every section is load-bearing.

---

## What Aphelion Is

Aphelion is an AI agent marketplace and infrastructure platform. Three types of users:

1. **API Owners** — have an existing API, upload an OpenAPI spec, Aphelion turns it into callable tools, they earn money per API call from other developers
2. **Agent Developers** — write agents that orchestrate multiple tools to do a job, deploy to Aphelion cloud, earn per agent execution
3. **Buyers** — discover and subscribe to agents in the marketplace, pay per execution via credit packs

**The CLI is the SDK.** A developer should never write HTTP clients, auth headers, session management, or memory plumbing. They write business logic. The CLI wraps everything. When `aphelion agent run agent.py` executes, the CLI injects auth, handles memory, calls tools, logs everything. The developer's code just imports `from aphelion import tools, memory` and calls functions.

**The target experience:** A vibecoder opens Claude and says "use Aphelion to create a review management agent." Claude installs the CLI, runs `aphelion auth login`, runs `aphelion agent init`, writes the agent logic, subscribes to tools, runs it locally, deploys it, and invokes it — entirely from the terminal. No console required for any of this.

---

## Platform API

```
Base URL:       https://api.aphl.ai
Auth0 Domain:   dev-ay0w6h2rrecsopt8.us.auth0.com
Auth0 Audience: https://aphelion-gateway.com
Auth0 Client:   UbXxpQBSr9AsqpS2ln2jzmsmamromaFC
Docs:           https://api.aphl.ai/docs
Console:        https://console.aphl.ai
```

**CRITICAL:** The existing CLI hardcodes `https://api.aphelion.exmplr.ai` as the base URL. This is wrong. Change it to `https://api.aphl.ai` everywhere — default config, help text, error messages, README. Do this first before anything else.

---

## Authentication Model

Three auth contexts exist. The CLI must handle all three transparently.

### 1. Human (Auth0 PKCE)
Used for anything that requires account access: managing agents, viewing history, billing, publishing APIs.

```bash
aphelion auth login
# Opens browser → Auth0 authorization URL
# Listens on localhost callback
# Exchanges code for access + refresh tokens
# Stores in ~/.aphelion/config.yaml
# Prints: "Logged in as jyothi@centaurisystems.io"
```

Token refresh must be automatic and silent. If access token is expired and refresh token is valid, refresh without prompting user. Only re-prompt if refresh token is expired.

### 2. Agent credentials (client_credentials grant)
Used when an agent runs. Client ID + secret → scoped agent JWT. Developer never writes this.

When the CLI executes an agent (`aphelion agent run` or `aphelion invoke`), it reads `.aphelion/config.yaml`, extracts `agent.client_id` and `agent.client_secret`, POSTs to `POST /auth/agent/token` to get a short-lived JWT, and injects it into every tool call and memory operation. This is completely invisible to the developer.

### 3. API key (legacy)
Keep working. Do not remove. Some existing integrations use it.

### Auth commands

```
aphelion auth login      # Auth0 browser PKCE flow
aphelion auth logout     # Clear stored tokens
aphelion auth profile    # Show current user: name, email, account ID, plan
aphelion auth status     # Show which context is active and token expiry
aphelion auth token      # Print current bearer token (for debugging/scripting)
```

Remove `aphelion auth register` and `aphelion auth oauth` — replaced by Auth0 flow.

---

## Config File Hierarchy

```
~/.aphelion/config.yaml      # Global: auth tokens, preferred API URL, output format
./.aphelion/config.yaml      # Project: agent credentials, tool subscriptions, deployment info
./.aphelion/.gitignore       # Always created, always contains: config.yaml, .env, *.log
./.env                       # Developer-managed secrets (TWILIO_PHONE_NUMBER etc)
```

Project config takes precedence over global. Merge on read, project wins on conflict.

**Never log or print `client_secret` or auth tokens.** Log only token prefix (first 8 chars + "...").

Global config structure:
```yaml
api_url: https://api.aphl.ai
output: table
auth:
  access_token: eyJ...
  refresh_token: eyJ...
  expires_at: "2026-03-13T10:00:00Z"
  user_email: jyothi@centaurisystems.io
  account_id: acc_xxxx
```

Project config structure:
```yaml
name: review-management-agent
description: Sends review requests to patients after visits
version: 1.0.0
language: python

agent:
  id: agt_xxxxxxxxxxxxxxxxxxxx
  client_id: agt_KFnG9Ad2r9LsCrJdLu9k
  client_secret: agt_secret_XXXXXXXXXXXXXXXX

gateway:
  api_url: https://api.aphl.ai

tools:
  subscribed:
    - twilio
    - sendgrid

execution:
  timeout: 30s
  memory_auto_save: true
  max_retries: 2

deployment:
  status: not_deployed
  endpoint: null
  region: us-central1
  last_deployed: null

schedule:
  cron: null
  enabled: false

logging:
  level: info
  file: agent.log
```

---

## Command Reference

Implement every command listed. Group them as shown. Existing commands that already work correctly should not be broken.

### `aphelion agent` — Agent Development

#### `aphelion agent init`
Most important command. Creates a complete, working agent project. Run interactively.

```
Prompts:
  Agent name (kebab-case):     review-management-agent
  Description:                  Sends review requests to patients after visits
  Language [python/node/go]:    python
  Subscribe to tools (search):  > twilio, sendgrid
  Create agent identity now? [Y/n]

Creates:
  review-management-agent/
  ├── agent.py                    # Main logic — developer edits this
  ├── requirements.txt            # aphelion-sdk plus user-selected tool deps
  ├── package.json                # (Node only)
  ├── .aphelion/
  │   ├── config.yaml            # Agent config including credentials (git-ignored)
  │   ├── agent.json             # Input/output manifest
  │   └── .gitignore             # Excludes config.yaml, .env, *.log
  ├── .env.example               # Template showing required secrets
  ├── README.md                  # How to run, deploy, invoke this agent
  └── tests/
      └── test_agent.py          # Test scaffold with example inputs
```

Generated `agent.py` must be complete and immediately runnable. Not a stub. It must:
- Import from the `aphelion` SDK
- Have a decorated `main(input)` async function
- Include a realistic example for the agent type (review management = Twilio SMS + SendGrid email)
- Include memory read and write examples
- Include `agent.env()` calls for secrets
- Include the `if __name__ == "__main__": agent.run_local()` block
- Have inline comments explaining every SDK call

```python
"""
{agent_name}
{description}

Deploy:      aphelion deploy
Run locally: aphelion agent run agent.py
Schedule:    aphelion agent run agent.py --cron "0 9 * * *"
Invoke:      aphelion invoke {agent_name} --input '{...}'
"""

from aphelion import Agent, tools, memory

agent = Agent()

@agent.run
async def main(input: dict) -> dict:
    """
    Input:
      patient_name: str  — required
      contact: str       — required, phone (+1...) or email
      visit_type: str    — optional

    Output:
      status: "sent" | "failed"
      channel: "sms" | "email"
      timestamp: ISO8601
      error: str | null
    """

    patient_name = input.get("patient_name")
    contact = input.get("contact")
    visit_type = input.get("visit_type", "your recent visit")

    if not patient_name or not contact:
        return {"status": "failed", "error": "patient_name and contact are required"}

    # Check if we've contacted this patient recently (avoid duplicate requests)
    recent = await memory.get(f"last_request:{contact}")
    if recent and recent.get("sent_within_days", 0) < 7:
        agent.log(f"Skipping {contact} — already contacted within 7 days")
        return {"status": "skipped", "reason": "recently_contacted"}

    message = (
        f"Hi {patient_name}, thank you for {visit_type}. "
        f"We'd love your feedback — it only takes 30 seconds: "
        f"{agent.env('REVIEW_LINK', default='https://g.page/r/your-review-link')}"
    )

    is_phone = contact.startswith("+") or contact.replace("-","").replace(" ","").isdigit()

    if is_phone:
        result = await tools.execute("twilio.messaging.create_message", {
            "To": contact,
            "Body": message,
            "From": agent.env("TWILIO_PHONE_NUMBER")
        })
        channel = "sms"
    else:
        result = await tools.execute("sendgrid.mail.send", {
            "personalizations": [{"to": [{"email": contact}]}],
            "from": {"email": agent.env("SENDGRID_FROM_EMAIL")},
            "subject": f"How was your experience, {patient_name}?",
            "content": [{"type": "text/plain", "value": message}]
        })
        channel = "email"

    # Write to memory — persists across executions, scoped to this agent
    await memory.set(f"last_request:{contact}", {
        "patient": patient_name,
        "channel": channel,
        "visit_type": visit_type,
        "sent_within_days": 0
    }, ttl="7d")

    return {
        "status": "sent" if result.get("success") else "failed",
        "channel": channel,
        "timestamp": result.get("timestamp"),
        "error": result.get("error")
    }


if __name__ == "__main__":
    # Run locally: python agent.py
    # Or:          aphelion agent run agent.py --input '{"patient_name": "Jane", "contact": "+15551234567"}'
    agent.run_local()
```

Generated `.env.example`:
```
# Copy to .env and fill in your values
# .env is git-ignored — never commit secrets

TWILIO_PHONE_NUMBER=+1XXXXXXXXXX
SENDGRID_FROM_EMAIL=noreply@yourdomain.com
REVIEW_LINK=https://g.page/r/your-review-link
```

Generated `agent.json`:
```json
{
  "name": "{agent_name}",
  "description": "{description}",
  "version": "1.0.0",
  "inputs": {
    "patient_name": {"type": "string", "required": true},
    "contact": {"type": "string", "required": true, "description": "Phone (+E.164) or email"},
    "visit_type": {"type": "string", "required": false}
  },
  "outputs": {
    "status": {"type": "string", "enum": ["sent", "failed", "skipped"]},
    "channel": {"type": "string", "enum": ["sms", "email"]},
    "timestamp": {"type": "string", "format": "date-time"},
    "error": {"type": "string", "nullable": true}
  },
  "pricing": {
    "model": "per_execution",
    "price": 0.01,
    "currency": "USD"
  },
  "visibility": "private"
}
```

#### `aphelion agent run [file]`
Run agent locally. Existing behavior is mostly correct. Updates needed:
- Add `--input` flag: accepts inline JSON passed to agent as `input` dict
- Add `--input-file` flag: reads JSON from file
- Before execution, silently exchange agent credentials for JWT and inject into runtime
- After execution, print formatted output and execution time
- If `--watch` flag: stream logs line by line as they're emitted

```bash
aphelion agent run agent.py --input '{"patient_name": "Jane", "contact": "+15551234567"}'
aphelion agent run agent.py --input-file ./test-input.json
aphelion agent run agent.py --cron "0 9 * * MON-FRI"
aphelion agent run agent.py --daemon
aphelion agent run agent.py --verbose
```

Keep existing cron and daemon behavior.

---

### `aphelion agents` — Agent Identity Management

These mirror the Agents page in the console. All require account-level auth.

```bash
aphelion agents create --name <name> --description <desc>
# Creates agent identity on platform
# Returns client_id and client_secret (shown ONCE)
# If inside an agent project dir, offers to save to .aphelion/config.yaml

aphelion agents list
# Table: Name | ID | Status | Created | Last Active

aphelion agents get <name-or-id>
# Full details: credentials prefix, tool subscriptions, memory stats, deployment status

aphelion agents update <name-or-id> --description <desc>

aphelion agents rotate-secret <name-or-id>
# Prompts: "This will immediately invalidate the current secret. Continue? [y/N]"
# Returns new secret ONCE
# Updates .aphelion/config.yaml if this is the current project agent

aphelion agents suspend <name-or-id>
# Blocks all executions. Existing deployments return 503.

aphelion agents activate <name-or-id>

aphelion agents delete <name-or-id>
# Prompts for confirmation, requires typing agent name

aphelion agents inspect <name-or-id>
# Shows: tool subscriptions, memory entry count, permission grants, deployment status, recent executions
```

---

### `aphelion agents grant/revoke` — Permissions Between Agents

RBAC so agents can read each other's memory. Enables multi-agent workflows.

```bash
aphelion agents grant \
  --from <grantee-agent> \
  --to <resource-agent> \
  --actions read

aphelion agents grant \
  --from researcher-agent \
  --to writer-agent \
  --actions read,write \
  --expires "2026-12-31"

aphelion agents permissions <name-or-id>
# Who has been granted access to this agent's memory

aphelion agents grants <name-or-id>
# What this agent has been granted access to

aphelion agents revoke \
  --from <grantee-agent> \
  --to <resource-agent>
# Prompts for confirmation
```

---

### `aphelion deploy` — Deploy to Aphelion Cloud

This is the second most important command after `agent init`. Run from inside an agent project directory.

```bash
aphelion deploy
aphelion deploy --agent my-agent-name     # Override name
aphelion deploy --public                  # List in marketplace after deploy
aphelion deploy --dry-run                 # Validate only, do not deploy
aphelion deploy --region us-central1      # Target region (default: us-central1)
```

What deploy does step by step:
1. Read `.aphelion/config.yaml` and `agent.json` — validate both exist
2. Validate agent code syntax (run `python -m py_compile agent.py` for Python)
3. Check tool subscriptions are active
4. Package: create tarball of agent code + dependencies
5. Upload to `POST /v2/agents/{agent-id}/deploy` with multipart
6. Poll deployment status until complete (with spinner)
7. Run health check: invoke with empty input, expect non-500 response
8. Update `.aphelion/config.yaml` with endpoint URL and deployment timestamp
9. Print endpoint and invocation example

Output format:
```
Validating agent...             ✓
Packaging dependencies...       ✓ (2.3 MB)
Uploading to Aphelion...        ✓
Provisioning cloud function...  ✓
Running health check...         ✓

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Agent deployed: review-management-agent
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Endpoint:  https://api.aphl.ai/v2/agents/agt_.../invoke
  Console:   https://console.aphl.ai/agents/review-management-agent
  Status:    active

  Invoke:
    aphelion invoke review-management-agent \
      --input '{"patient_name": "Jane", "contact": "+15551234567"}'

  Or via curl:
    curl -X POST https://api.aphl.ai/v2/agents/agt_.../invoke \
      -H "Authorization: Bearer $(aphelion auth token)" \
      -H "Content-Type: application/json" \
      -d '{"patient_name": "Jane", "contact": "+15551234567"}'
```

---

### `aphelion deployments` — Manage Deployed Agents

```bash
aphelion deployments list
# Table: Agent | Status | Endpoint | Region | Last Deploy | Executions (24h)

aphelion deployments status
# Status of current project's deployment

aphelion deployments logs <agent-name>
# Recent logs from deployed agent (last 100 lines)

aphelion deployments logs <agent-name> --follow
# Stream live logs (tail -f equivalent, Ctrl+C to stop)

aphelion deployments history <agent-name>
# Recent executions: timestamp, input summary, output summary, duration, status
aphelion deployments history <agent-name> --limit 50 --status failed

aphelion deployments rollback <agent-name>
# Rollback to previous deployment version

aphelion deployments stop <agent-name>
# Stop serving requests. Keeps agent registered.

aphelion deployments redeploy <agent-name>
# Re-run deploy from current code

aphelion deployments delete <agent-name>
# Prompts confirmation. Stops serving, removes endpoint.
```

---

### `aphelion invoke` — Call a Deployed Agent

```bash
aphelion invoke review-management-agent \
  --input '{"patient_name": "Jane", "contact": "+15551234567"}'

aphelion invoke review-management-agent \
  --input-file ./test-input.json

aphelion invoke review-management-agent \
  --input '{"patient_name": "Jane"}' \
  --watch
# Streams execution logs in real time, then prints result

# Invoke another developer's public agent from marketplace
aphelion invoke exmplr/clinical-search \
  --input '{"query": "NSCLC phase 3 trials"}'
```

Print result as formatted JSON by default. `--output table` for summary view.

---

### `aphelion memory` — Agent Memory (Updated)

All memory is agent-scoped. If running inside an agent project directory, agent is inferred from `.aphelion/config.yaml`. Otherwise `--agent` is required.

```bash
aphelion memory list
aphelion memory list --agent review-management-agent
aphelion memory list --limit 50 --search "patient"

aphelion memory get <key>
aphelion memory get "last_request:+15551234567"

aphelion memory set <key> <value>
aphelion memory set "config:review_link" "https://g.page/r/..."
aphelion memory set "config:review_link" "https://g.page/r/..." --ttl 30d

aphelion memory delete <key>

aphelion memory search <query>
# Semantic search across agent's memory namespace
aphelion memory search "patient contact history" --threshold 0.7

aphelion memory stats
# Total entries, size, oldest entry, most recent entry

aphelion memory clear
# Prompts: "Delete all 47 memory entries for review-management-agent? [y/N]"
aphelion memory clear --session <session-id>

# Cross-agent memory read (requires permission)
aphelion memory get <key> --from researcher-agent
```

---

### `aphelion tools` — Tool Discovery and Management (Updated)

```bash
aphelion tools search "sms messaging"
aphelion tools search "email delivery"
aphelion tools marketplace
aphelion tools marketplace --category communication
aphelion tools marketplace --category data
aphelion tools marketplace --category ai
aphelion tools marketplace --free
aphelion tools marketplace --paid

# Subscribe current agent to a tool
aphelion tools subscribe twilio
aphelion tools subscribe sendgrid

# Unsubscribe
aphelion tools unsubscribe twilio

# List tools current agent is subscribed to
aphelion tools list

# Describe a specific tool's parameters, schema, examples
aphelion tools describe twilio.messaging.create_message
# Shows: description, required params, optional params, example call, pricing

# Execute a tool directly (for testing, not inside agent code)
aphelion tools try \
  --tool twilio.messaging.create_message \
  --params '{"To": "+15551234567", "Body": "Test", "From": "+15559876543"}'

aphelion tools try \
  --tool sendgrid.mail.send \
  --params '{"to": "test@example.com"}' \
  --dry-run
# Validates params without executing
```

---

### `aphelion registry` — Publish APIs as Tools (Updated)

For API owners publishing their APIs to the marketplace.

```bash
aphelion registry add-openapi --file ./openapi.json
aphelion registry add-openapi --url https://api.example.com/openapi.json
aphelion registry add-openapi --file ./openapi.json \
  --name "My API" \
  --description "Does something useful" \
  --base-url "https://api.example.com"

aphelion registry list
aphelion registry my-services    # Alias for list

aphelion registry get <service-id>

aphelion registry update <service-id> --price 0.002
aphelion registry update <service-id> --description "Updated description"

# Visibility
aphelion registry publish <service-id> --visibility public
aphelion registry publish <service-id> --visibility private

aphelion registry unpublish <service-id>

aphelion registry delete <service-id>

# Earnings for published APIs
aphelion registry earnings
aphelion registry earnings --service <service-id>
aphelion registry earnings --last 30d
```

---

### `aphelion env` — Secret Management for Deployed Agents

```bash
# Set secret for deployed agent (stored server-side, injected at runtime)
aphelion env set TWILIO_PHONE_NUMBER "+15551234567"
aphelion env set SENDGRID_FROM_EMAIL "noreply@medgenie.com"

# List keys (never show values)
aphelion env list
# Output: TWILIO_PHONE_NUMBER, SENDGRID_FROM_EMAIL, REVIEW_LINK

# Delete a secret
aphelion env delete TWILIO_PHONE_NUMBER

# Pull all env vars from deployed agent to local .env (shows masked values)
aphelion env pull

# Push local .env to deployed agent env vars
aphelion env push
# Prompts confirmation, shows which keys will be set/updated/deleted
```

---

### `aphelion schedule` — Scheduling for Deployed Agents

```bash
aphelion schedule set review-management-agent --cron "0 9 * * MON-FRI"
# Prints: "Agent will run at 9:00 AM Monday-Friday (America/Los_Angeles)"

aphelion schedule get review-management-agent
# Shows: cron expression, next 5 run times, enabled/disabled

aphelion schedule disable review-management-agent
aphelion schedule enable review-management-agent
aphelion schedule remove review-management-agent
```

Keep existing `aphelion agent run agent.py --cron "..."` for local scheduling. These are different — `aphelion schedule` operates on the deployed cloud function, `agent run --cron` runs on the local machine.

---

### `aphelion analytics` — Usage and Earnings

```bash
aphelion analytics
# Dashboard for current account: executions, spend, earnings, top agents

aphelion analytics --agent review-management-agent
# Agent-specific: execution count, success rate, avg duration, unique callers

aphelion analytics tools
# Tool usage breakdown: which tools called, how often, total spend

aphelion analytics executions
aphelion analytics executions --agent review-management-agent
aphelion analytics executions --status failed
aphelion analytics executions --last 7d
aphelion analytics executions --last 30d

aphelion analytics earnings
aphelion analytics earnings --last 30d
# For API/agent publishers: calls received, revenue, top consumers
```

---

### `aphelion status` — Project Status

Run from inside an agent project directory. Shows everything at a glance.

```
aphelion status

──────────────────────────────────────────
  review-management-agent
──────────────────────────────────────────
  Agent ID:     agt_KFnG9Ad2r9LsCrJdLu9k
  Status:       deployed ✓
  Endpoint:     https://api.aphl.ai/v2/agents/agt_.../invoke
  Region:       us-central1
  Last deploy:  2 hours ago

  Tools:        twilio ✓  sendgrid ✓
  Schedule:     0 9 * * MON-FRI (enabled)
  Env vars:     TWILIO_PHONE_NUMBER, SENDGRID_FROM_EMAIL, REVIEW_LINK

  Memory:       47 entries
  Executions:   128 total  |  124 success (96.9%)  |  4 failed
  Last run:     14 minutes ago (success, 1.2s)
──────────────────────────────────────────
  Console: https://console.aphl.ai/agents/review-management-agent
```

---

### `aphelion mcp serve` — MCP Server (Critical for Claude Integration)

Expose all Aphelion CLI functionality as MCP tools so Claude and other AI assistants can use Aphelion directly without a user typing commands.

```bash
aphelion mcp serve
# Starts MCP server on stdio (default) or --port for SSE transport
# Claude Desktop, Claude.ai, and compatible AI tools connect to this
```

MCP tools to expose:

```
aphelion_auth_status          → Is user authenticated?
aphelion_agent_init           → Initialize new agent project
aphelion_agent_run            → Run agent locally with input
aphelion_agents_list          → List all agents on account
aphelion_agents_create        → Create new agent identity
aphelion_agents_inspect       → Get agent details
aphelion_deploy               → Deploy current agent to cloud
aphelion_invoke               → Invoke a deployed agent
aphelion_deployments_status   → Get deployment status
aphelion_deployments_logs     → Get recent logs
aphelion_memory_get           → Read a memory key
aphelion_memory_set           → Write a memory key
aphelion_memory_search        → Semantic search across memory
aphelion_memory_list          → List memory entries
aphelion_tools_search         → Search marketplace tools
aphelion_tools_subscribe      → Subscribe agent to a tool
aphelion_tools_describe       → Get tool parameter schema
aphelion_tools_try            → Execute a tool directly
aphelion_env_set              → Set environment variable for agent
aphelion_env_list             → List environment variable keys
aphelion_analytics            → Get usage analytics
aphelion_status               → Get current project status
```

Claude Desktop config (print this with `aphelion mcp config`):
```json
{
  "mcpServers": {
    "aphelion": {
      "command": "aphelion",
      "args": ["mcp", "serve"],
      "env": {}
    }
  }
}
```

When this MCP server is running, a user can open Claude Desktop and say "use Aphelion to create a review management agent" and Claude will call these tools to init, write, deploy, and invoke the agent without the user touching the terminal.

---

### Utility Commands

```bash
aphelion quickstart
# Interactive tutorial: auth → init → subscribe tools → run → deploy

aphelion whoami
# Show: email, account ID, plan, agent count, API count

aphelion open
# Open console.aphl.ai in browser
aphelion open agents
aphelion open marketplace
aphelion open history
aphelion open docs

aphelion version
# Show CLI version, API version, Go version

aphelion completion bash|zsh|fish|powershell
# Shell completion scripts (keep existing behavior)
```

---

## Aphelion Python SDK

The SDK is imported inside agent code. It must exist as a Python package. Options:
1. Publish to PyPI as `aphelion-sdk`
2. Bundle with CLI and install at `aphelion agent init` time via `pip install`

Minimum viable SDK surface:

```python
# aphelion/__init__.py
from .agent import Agent
from . import tools, memory
```

```python
# aphelion/agent.py
class Agent:
    def run(self, func):
        """Decorator that registers the agent's main function"""
    
    def run_local(self):
        """Called by __main__ for local testing. Reads --input flag or prompts."""
    
    def env(self, key, default=None):
        """Reads from .env file. Raises clear error if missing and no default."""
    
    def log(self, message, level="info"):
        """Writes to agent.log and stdout during run."""
```

```python
# aphelion/tools.py
async def execute(tool_name: str, params: dict) -> dict:
    """
    Calls POST /v1/agents/{session_id}/tools/{tool_name}/execute
    Auth header injected from runtime context (set by CLI before agent starts)
    Returns response body as dict
    Raises ToolError with clear message on failure
    """

async def list() -> list:
    """Returns list of tools current agent is subscribed to"""
```

```python
# aphelion/memory.py
async def get(key: str) -> dict | None:
    """GET /v2/agents/{agent_id}/memory/{key}. Returns None if not found."""

async def set(key: str, value: dict, ttl: str = None) -> None:
    """PUT /v2/agents/{agent_id}/memory/{key}. ttl format: "7d", "24h", "30m"."""

async def delete(key: str) -> None:
    """DELETE /v2/agents/{agent_id}/memory/{key}"""

async def search(query: str, threshold: float = 0.7) -> list:
    """POST /v2/agents/{agent_id}/memory/search"""
```

The CLI injects the agent ID, session ID, and auth token into the runtime via environment variables before executing the agent code. The SDK reads these from environment — the developer never sees them.

```
APHELION_AGENT_ID=agt_xxx
APHELION_SESSION_ID=ses_xxx  
APHELION_API_TOKEN=eyJ...
APHELION_API_URL=https://api.aphl.ai
```

Node.js SDK (`@aphelion/sdk`) should be equivalent. Include scaffold for both when language is selected at `agent init`.

---

## Error Messages

Every error must be actionable. Never print a raw HTTP status code alone.

```
# WRONG
Error: 401 Unauthorized

# RIGHT
Session expired. Run: aphelion auth login

# WRONG  
Error: 404 Not Found

# RIGHT
Agent "review-agent" not found.
List your agents: aphelion agents list

# WRONG
Tool execution failed: 400

# RIGHT
Tool "twilio.messaging.create_message" failed.
Reason: "To" must be in E.164 format (e.g. +15551234567)
Docs:   aphelion tools describe twilio.messaging.create_message

# WRONG
Error: agent not deployed

# RIGHT
Agent is not deployed yet.
Deploy with: aphelion deploy
```

---

## End-to-End Test Flow

Before any PR is complete, verify this entire flow works without errors:

```bash
# Install
brew install aphelion

# Auth
aphelion auth login
aphelion whoami
# Expected: shows email and account ID

# Init
mkdir review-agent && cd review-agent
aphelion agent init
# Expected: full project created, agent identity created on platform

# Tool setup
aphelion tools subscribe twilio
aphelion tools subscribe sendgrid
aphelion tools list
# Expected: twilio and sendgrid listed as subscribed

# Local run
cp .env.example .env
# (edit .env with real values)
aphelion agent run agent.py \
  --input '{"patient_name": "Jane Smith", "contact": "+15551234567"}'
# Expected: agent executes, SMS sent or mocked, result printed

# Memory
aphelion memory list
# Expected: entry from previous run exists
aphelion memory get "last_request:+15551234567"
# Expected: shows patient name, channel, timestamp

# Deploy
aphelion env set TWILIO_PHONE_NUMBER "+15551234567"
aphelion env set SENDGRID_FROM_EMAIL "noreply@test.com"
aphelion deploy
# Expected: endpoint URL printed

# Invoke
aphelion invoke review-management-agent \
  --input '{"patient_name": "Test Patient", "contact": "test@example.com"}'
# Expected: success response

# Logs
aphelion deployments logs review-management-agent
# Expected: shows execution logs from invoke

# Status
aphelion status
# Expected: shows all details including endpoint, memory count, last run

# MCP
aphelion mcp serve &
# Expected: starts without error, MCP tools accessible

# Analytics
aphelion analytics --agent review-management-agent
# Expected: shows execution stats including the test runs
```

---

## Things Not To Build In This Update

Do not scope creep into these. They are future work:

- STELLA agent-to-agent communication protocol
- Real-time WebSocket streaming from deployed agents
- Multi-region deployment targeting
- Custom domain mapping for agent endpoints  
- Team/org management (invite members, roles)
- Billing and payout CLI commands (use console for now)
- Marketplace discovery of other developers' published agents (console only for now)
- Agent versioning and A/B testing

---

## Implementation Order

Do this in order. Do not skip ahead.

1. **Fix API URL** — change `api.aphelion.exmplr.ai` to `api.aphl.ai` everywhere
2. **Update auth** — Auth0 PKCE flow replacing username/password
3. **Add `aphelion agents` commands** — create, list, get, rotate-secret, suspend, activate, delete, inspect
4. **Add agent permission commands** — grant, revoke, permissions, grants
5. **Update `aphelion agent init`** — full project scaffold with complete working agent code
6. **Update memory commands** — agent-scoped, `--agent` flag, new get/set/delete
7. **Add `aphelion deploy`** — the most important new command
8. **Add `aphelion deployments`** — list, status, logs, history, stop, redeploy, rollback
9. **Add `aphelion invoke`**
10. **Add `aphelion env`**
11. **Add `aphelion schedule`**
12. **Update `aphelion tools`** — subscribe, unsubscribe, marketplace, list
13. **Update analytics** — agent-scoped filters
14. **Add `aphelion status`**
15. **Add `aphelion mcp serve`** — highest leverage, do not skip
16. **Update README** — full new command reference, quickstart flow

Run the end-to-end test flow after step 8 and again at the end.
