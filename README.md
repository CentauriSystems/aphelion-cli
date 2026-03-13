# Aphelion CLI

The SDK for Aphelion — an AI agent marketplace and infrastructure platform.

Build agents that orchestrate API tools, deploy them to the cloud, and earn per execution when others use them in the marketplace. The CLI handles auth, memory, tool calls, and deployment so you write only business logic.

Agent code can be written in **Python**, **Node.js**, or **Go**.

## Installation

```bash
# Homebrew (macOS/Linux)
brew tap exmplr-ai/aphelion
brew install aphelion

# Or install with Go
go install github.com/Exmplr-AI/aphelion-cli@latest
```

## Quickstart

```bash
# 1. Authenticate via browser
aphelion auth login

# 2. Create a new agent project
aphelion agent init

# 3. Subscribe to tools your agent needs
aphelion tools subscribe twilio
aphelion tools subscribe sendgrid

# 4. Configure secrets
cp .env.example .env
# Edit .env with your API keys

# 5. Run locally
aphelion agent run agent.py \
  --input '{"patient_name": "Jane", "contact": "+15551234567"}'

# 6. Deploy to Aphelion Cloud
aphelion deploy
# => Agent deployed: review-management-agent
# => Endpoint: https://api.aphl.ai/v2/agents/agt_.../invoke
# => Console: https://console.aphl.ai/agents/review-management-agent
```

### What your agent code looks like

`aphelion agent init` generates a complete, runnable agent. Here's what the Python code looks like:

```python
from aphelion import Agent, tools, memory

agent = Agent()

@agent.run
async def main(input: dict) -> dict:
    patient_name = input.get("patient_name")
    contact = input.get("contact")

    # Check memory — avoid contacting the same patient twice in a week
    recent = await memory.get(f"last_request:{contact}")
    if recent and recent.get("sent_within_days", 0) < 7:
        return {"status": "skipped", "reason": "recently_contacted"}

    # Call a tool — Aphelion handles auth and routing
    result = await tools.execute("twilio.messaging.create_message", {
        "To": contact,
        "Body": f"Hi {patient_name}, we'd love your feedback.",
        "From": agent.env("TWILIO_PHONE_NUMBER")
    })

    # Write to memory — persists across executions, scoped to this agent
    await memory.set(f"last_request:{contact}", {
        "patient": patient_name, "sent_within_days": 0
    }, ttl="7d")

    return {"status": "sent", "channel": "sms"}

if __name__ == "__main__":
    agent.run_local()
```

You import `tools` and `memory`, call them as async functions, and the CLI injects auth, routing, and persistence at runtime. No HTTP clients, no headers, no session management.

## Command Reference

### Authentication

| Command | Description |
|---------|-------------|
| `aphelion auth login` | Authenticate via browser (Auth0 PKCE) |
| `aphelion auth logout` | Clear stored tokens |
| `aphelion auth profile` | Show current user: name, email, account ID, plan |
| `aphelion auth status` | Show active auth context and token expiry |
| `aphelion auth token` | Print current bearer token (for scripting) |

### Agent Development

| Command | Description |
|---------|-------------|
| `aphelion agent init` | Create a complete, runnable agent project interactively |
| `aphelion agent run [file]` | Run agent locally |
| `aphelion agent run [file] --input '{...}'` | Run with inline JSON input |
| `aphelion agent run [file] --input-file data.json` | Run with input from file |
| `aphelion agent run [file] --cron "0 9 * * *"` | Run on a local cron schedule |
| `aphelion agent run [file] --daemon` | Run as a background daemon |
| `aphelion agent run [file] --verbose` | Run with verbose output |

### Agent Identity Management

| Command | Description |
|---------|-------------|
| `aphelion agents create` | Create a new agent identity on the platform |
| `aphelion agents list` | List all agents on your account |
| `aphelion agents get <name-or-id>` | Show full agent details |
| `aphelion agents update <name-or-id>` | Update agent metadata |
| `aphelion agents rotate-secret <name-or-id>` | Rotate agent client secret |
| `aphelion agents suspend <name-or-id>` | Suspend an agent (blocks executions) |
| `aphelion agents activate <name-or-id>` | Re-activate a suspended agent |
| `aphelion agents delete <name-or-id>` | Delete an agent permanently |
| `aphelion agents inspect <name-or-id>` | Show subscriptions, memory, permissions, executions |

### Agent Permissions

| Command | Description |
|---------|-------------|
| `aphelion agents grant --from <agent> --to <agent> --actions read` | Grant cross-agent memory access |
| `aphelion agents revoke --from <agent> --to <agent>` | Revoke a permission grant |
| `aphelion agents permissions <name-or-id>` | Show who has access to this agent's memory |
| `aphelion agents grants <name-or-id>` | Show what this agent has been granted access to |

### Deployment

| Command | Description |
|---------|-------------|
| `aphelion deploy` | Deploy current project to Aphelion Cloud |
| `aphelion deploy --public` | Deploy and list in the marketplace |
| `aphelion deploy --dry-run` | Validate without deploying |
| `aphelion deploy --region us-central1` | Deploy to a specific region |

### Deployment Management

| Command | Description |
|---------|-------------|
| `aphelion deployments list` | List all deployed agents |
| `aphelion deployments status` | Status of current project's deployment |
| `aphelion deployments logs <agent>` | Recent logs from deployed agent |
| `aphelion deployments logs <agent> --follow` | Stream live logs |
| `aphelion deployments history <agent>` | Recent execution history |
| `aphelion deployments rollback <agent>` | Rollback to previous version |
| `aphelion deployments stop <agent>` | Stop serving requests |
| `aphelion deployments redeploy <agent>` | Re-deploy from current code |
| `aphelion deployments delete <agent>` | Remove deployment entirely |

### Invocation

| Command | Description |
|---------|-------------|
| `aphelion invoke <agent> --input '{...}'` | Call a deployed agent |
| `aphelion invoke <agent> --input-file data.json` | Call with input from file |
| `aphelion invoke <agent> --watch` | Stream execution logs then print result |

### Memory

All memory is agent-scoped. Inside a project directory, the agent is inferred automatically.

| Command | Description |
|---------|-------------|
| `aphelion memory list` | List memory entries |
| `aphelion memory get <key>` | Read a memory entry |
| `aphelion memory set <key> <value>` | Write a memory entry |
| `aphelion memory set <key> <value> --ttl 7d` | Write with expiration |
| `aphelion memory delete <key>` | Delete a memory entry |
| `aphelion memory search <query>` | Semantic search across memory |
| `aphelion memory stats` | Show entry count and size |
| `aphelion memory clear` | Delete all entries (with confirmation) |
| `aphelion memory get <key> --from <agent>` | Read from another agent (requires permission) |

Use `--agent <name>` to target a specific agent from outside a project directory.

### Tools

| Command | Description |
|---------|-------------|
| `aphelion tools search "sms messaging"` | Search marketplace for tools |
| `aphelion tools marketplace` | Browse the tool marketplace |
| `aphelion tools marketplace --category communication` | Filter by category |
| `aphelion tools subscribe <tool>` | Subscribe current agent to a tool |
| `aphelion tools unsubscribe <tool>` | Unsubscribe from a tool |
| `aphelion tools list` | List tools current agent is subscribed to |
| `aphelion tools describe <tool>` | Show parameters, schema, examples, pricing |
| `aphelion tools try --tool <name> --params '{...}'` | Execute a tool directly |
| `aphelion tools try --tool <name> --params '{...}' --dry-run` | Validate params without executing |

### Environment Variables

Manage secrets for deployed agents. Values are stored server-side and injected at runtime.

| Command | Description |
|---------|-------------|
| `aphelion env set <KEY> <VALUE>` | Set a secret for the deployed agent |
| `aphelion env list` | List secret keys (values are never shown) |
| `aphelion env delete <KEY>` | Delete a secret |
| `aphelion env pull` | Pull deployed env vars to local `.env` (masked) |
| `aphelion env push` | Push local `.env` to deployed agent |

### Scheduling

Manage cron schedules for deployed agents (distinct from local `agent run --cron`).

| Command | Description |
|---------|-------------|
| `aphelion schedule set <agent> --cron "0 9 * * MON-FRI"` | Set a cloud schedule |
| `aphelion schedule get <agent>` | Show schedule and next run times |
| `aphelion schedule enable <agent>` | Enable schedule |
| `aphelion schedule disable <agent>` | Disable schedule |
| `aphelion schedule remove <agent>` | Remove schedule |

### Service Registry

For API owners publishing their APIs as tools in the marketplace.

| Command | Description |
|---------|-------------|
| `aphelion registry add-openapi --file spec.json` | Register API from OpenAPI spec |
| `aphelion registry add-openapi --url <url>` | Register API from URL |
| `aphelion registry list` | List all public services |
| `aphelion registry my-services` | List your registered services |
| `aphelion registry get <id>` | Get service details |
| `aphelion registry update <id>` | Update service metadata or pricing |
| `aphelion registry publish <id> --visibility public` | Publish to marketplace |
| `aphelion registry unpublish <id>` | Remove from marketplace |
| `aphelion registry delete <id>` | Delete a service |
| `aphelion registry earnings` | View revenue from published APIs |

### Analytics

| Command | Description |
|---------|-------------|
| `aphelion analytics` | Account dashboard: executions, spend, earnings |
| `aphelion analytics --agent <name>` | Agent-specific stats |
| `aphelion analytics tools` | Tool usage breakdown |
| `aphelion analytics sessions` | Session analytics |
| `aphelion analytics executions` | Execution history with filters |
| `aphelion analytics earnings` | Publisher revenue report |

### Project Status

```bash
aphelion status
```

Run from inside an agent project directory to see agent ID, deployment status, endpoint, tool subscriptions, schedule, env vars, memory stats, and execution summary at a glance.

### MCP Server

| Command | Description |
|---------|-------------|
| `aphelion mcp serve` | Start MCP server (stdio transport) |
| `aphelion mcp config` | Print Claude Desktop configuration |

### Utilities

| Command | Description |
|---------|-------------|
| `aphelion whoami` | Show email, account ID, plan, agent/API count |
| `aphelion open` | Open the Aphelion console in browser |
| `aphelion open agents` | Open agents page |
| `aphelion open marketplace` | Open marketplace |
| `aphelion open docs` | Open documentation |
| `aphelion quickstart` | Interactive tutorial walkthrough |
| `aphelion version` | Show CLI, API, and Go versions |
| `aphelion completion bash\|zsh\|fish\|powershell` | Generate shell completions |

## MCP Integration

Aphelion exposes all CLI functionality as MCP tools, so AI assistants like Claude can create, deploy, and invoke agents directly.

### Setup with Claude Desktop

Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "aphelion": {
      "command": "aphelion",
      "args": ["mcp", "serve"]
    }
  }
}
```

Or generate this config automatically:

```bash
aphelion mcp config
```

Once configured, you can tell Claude: "Use Aphelion to create a review management agent" and it will call the MCP tools to init, write, deploy, and invoke the agent without you touching the terminal.

### Hosted MCP (no install required)

Add to Claude Desktop or Claude.ai connector settings:

```
URL:  https://mcp.aphl.ai/mcp
Auth: Bearer YOUR_APHELION_API_KEY
```

All tools and agents in the Aphelion marketplace are immediately available to Claude. No CLI required.

## Configuration

### Global config: `~/.aphelion/config.yaml`

Stores auth tokens, API URL, and output preferences. Created on first `aphelion auth login`.

```yaml
api_url: https://api.aphl.ai
output: table
auth:
  access_token: eyJ...
  refresh_token: eyJ...
  expires_at: "2026-03-13T10:00:00Z"
  user_email: user@example.com
  account_id: acc_xxxx
```

### Project config: `.aphelion/config.yaml`

Created by `aphelion agent init` inside each agent project. Stores agent credentials, tool subscriptions, and deployment info. Git-ignored by default.

Project config takes precedence over global config on conflicting keys.

### Environment Variables

| Variable | Description |
|----------|-------------|
| `APHELION_API_URL` | Override API base URL (default: `https://api.aphl.ai`) |
| `APHELION_OUTPUT` | Output format: `json`, `yaml`, `table` |
| `APHELION_VERBOSE` | Enable verbose output |

### Global Flags

| Flag | Description |
|------|-------------|
| `--config` | Specify config file path |
| `--api-url` | Override API base URL |
| `--output, -o` | Output format: `json`, `yaml`, `table` |
| `--verbose, -v` | Enable verbose output |

## Links

- **API Docs:** https://api.aphl.ai/docs
- **Console:** https://console.aphl.ai
- **GitHub:** https://github.com/Exmplr-AI/aphelion-cli
- **Issues:** https://github.com/Exmplr-AI/aphelion-cli/issues

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
