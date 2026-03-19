package agent

// All templates for agent init scaffolding.
// Python/JS templates use raw strings with placeholder markers
// ({{.Name}}, {{.Description}}) that are replaced via strings.ReplaceAll.
// This avoids conflicts with Python/JS curly braces and Go's text/template.

const pythonAgentTemplate = `"""
{{.Name}}
{{.Description}}

Deploy:      aphelion deploy
Run locally: aphelion agent run agent.py
Schedule:    aphelion agent run agent.py --cron "0 9 * * *"
Invoke:      aphelion invoke {{.Name}} --input '{...}'
"""

from aphelion import Agent, tools, memory

agent = Agent()

@agent.run
async def main(input: dict) -> dict:
    """
    Process input and return results.

    Customize this function with your agent's logic.
    Use tools.execute() to call subscribed tools,
    and memory.get()/set() to persist data across runs.

    Args:
        input: Dict with your agent's input parameters

    Returns:
        Dict with your agent's output
    """
    result = {}

    # Example: Read from memory
    previous = await memory.get("last_run")
    if previous:
        agent.log(f"Previous run: {previous}")

    # ------------------------------------------------------------------
    # Add your agent logic here.
    #
    # Call tools you've subscribed to:
    #   data = await tools.execute("service.operation", {"param": "value"})
    #
    # Read secrets from env vars (set via: aphelion env set KEY value):
    #   api_key = agent.env("MY_API_KEY")
    #
    # Search the marketplace for tools:
    #   aphelion tools search "sms"
    #   aphelion tools subscribe twilio
    # ------------------------------------------------------------------

    result["input_received"] = input
    result["status"] = "success"

    # Save to memory — persists across executions, scoped to this agent
    await memory.set("last_run", {
        "input": input,
        "status": "success"
    })

    return result


if __name__ == "__main__":
    # Run locally: python agent.py
    # Or:          aphelion agent run agent.py --input '{"key": "value"}'
    agent.run_local()
`

const nodeAgentTemplate = `/**
 * {{.Name}}
 * {{.Description}}
 *
 * Deploy:      aphelion deploy
 * Run locally: aphelion agent run index.js
 * Schedule:    aphelion agent run index.js --cron "0 9 * * *"
 * Invoke:      aphelion invoke {{.Name}} --input '{...}'
 */

const { Agent, tools, memory } = require("@aphelion/sdk");

const agent = new Agent();

agent.run(async (input) => {
  const result = {};

  // Example: Read from memory
  const previous = await memory.get("last_run");
  if (previous) {
    agent.log(` + "`Previous run: ${JSON.stringify(previous)}`" + `);
  }

  // ------------------------------------------------------------------
  // Add your agent logic here.
  //
  // Call tools you've subscribed to:
  //   const data = await tools.execute("service.operation", { param: "value" });
  //
  // Read secrets from env vars (set via: aphelion env set KEY value):
  //   const apiKey = agent.env("MY_API_KEY");
  //
  // Search the marketplace for tools:
  //   aphelion tools search "sms"
  //   aphelion tools subscribe twilio
  // ------------------------------------------------------------------

  result.input_received = input;
  result.status = "success";

  // Save to memory — persists across executions, scoped to this agent
  await memory.set("last_run", {
    input,
    status: "success",
  });

  return result;
});

if (require.main === module) {
  // Run locally: node index.js
  // Or:          aphelion agent run index.js --input '{"key": "value"}'
  agent.runLocal();
}
`

const agentJSONTemplate = `{
  "name": "{{.Name}}",
  "description": "{{.Description}}",
  "version": "1.0.0",
  "inputs": {},
  "outputs": {
    "status": {"type": "string"}
  },
  "pricing": {
    "model": "per_execution",
    "price": 0.01,
    "currency": "USD"
  },
  "visibility": "private"
}
`

const envExampleTemplate = `# Copy to .env and fill in your values
# .env is git-ignored — never commit secrets

# Add your secrets here, e.g.:
# MY_API_KEY=your-key-here
#
# Access in agent code:  agent.env("MY_API_KEY")
# Set for deployed agent: aphelion env set MY_API_KEY your-key-here
`

const requirementsTemplate = `aphelion-sdk>=0.1.0

# Add your own dependencies below
`

const sdkInitTemplate = `from .agent import Agent
from . import tools, memory

__all__ = ["Agent", "tools", "memory"]
`

const sdkAgentTemplate = `"""
Aphelion SDK — Agent class
Bundled with your agent project. Do not publish separately.
"""

import asyncio
import json
import logging
import os
import sys
from datetime import datetime, timezone
from functools import wraps

try:
    from dotenv import load_dotenv
except ImportError:
    load_dotenv = None


class Agent:
    """Aphelion agent wrapper. Handles lifecycle, env, and local execution."""

    def __init__(self):
        self._main_func = None
        self._logger = logging.getLogger("aphelion.agent")
        handler = logging.StreamHandler(sys.stderr)
        handler.setFormatter(logging.Formatter("%(asctime)s [%(levelname)s] %(message)s"))
        if not self._logger.handlers:
            self._logger.addHandler(handler)
        self._logger.setLevel(logging.INFO)

        # Load .env when running locally
        if load_dotenv is not None:
            load_dotenv(override=False)

    def run(self, func):
        """Decorator that registers the agent's main async function."""
        self._main_func = func

        @wraps(func)
        async def wrapper(*args, **kwargs):
            return await func(*args, **kwargs)

        return wrapper

    def env(self, key: str, default=None) -> str:
        """Read a secret from .env / environment.

        When running locally, python-dotenv loads .env automatically.
        When deployed, the CLI injects env vars before execution.
        """
        value = os.environ.get(key)
        if value is not None:
            return value
        if default is not None:
            return default
        raise RuntimeError(
            f"Missing required secret: {key}. "
            f"Add it to .env or run: aphelion env set {key} <value>"
        )

    def log(self, message: str, level: str = "info"):
        """Log a message with timestamp."""
        log_func = getattr(self._logger, level.lower(), self._logger.info)
        log_func(message)

    def run_local(self):
        """Entry point for local execution via python agent.py or aphelion agent run."""
        if self._main_func is None:
            print("Error: No function registered. Use @agent.run to decorate your main function.")
            sys.exit(1)

        input_data = self._parse_local_input()

        self.log(f"Running locally with input: {json.dumps(input_data, default=str)[:200]}")

        result = asyncio.run(self._main_func(input_data))

        # Single-line JSON — executor captures last stdout line as result
        print(json.dumps(result, default=str))

    @staticmethod
    def _parse_local_input() -> dict:
        """Parse --input or --input-file from sys.argv, or prompt on stdin."""
        # Check APHELION_INPUT env var first (set by CLI's aphelion agent run)
        env_input = os.environ.get("APHELION_INPUT", "")
        if env_input:
            try:
                return json.loads(env_input)
            except json.JSONDecodeError as e:
                print(f"Error: Invalid JSON in APHELION_INPUT: {e}")
                sys.exit(1)

        args = sys.argv[1:]

        for i, arg in enumerate(args):
            if arg == "--input" and i + 1 < len(args):
                try:
                    return json.loads(args[i + 1])
                except json.JSONDecodeError as e:
                    print(f"Error: Invalid JSON for --input: {e}")
                    sys.exit(1)
            if arg == "--input-file" and i + 1 < len(args):
                try:
                    with open(args[i + 1]) as f:
                        return json.load(f)
                except (OSError, json.JSONDecodeError) as e:
                    print(f"Error: Could not read input file: {e}")
                    sys.exit(1)

        # No flags — check if stdin has data (non-interactive)
        if not sys.stdin.isatty():
            try:
                return json.load(sys.stdin)
            except json.JSONDecodeError as e:
                print(f"Error: Invalid JSON on stdin: {e}")
                sys.exit(1)

        # Interactive prompt
        print("Enter input JSON (then press Enter):")
        try:
            line = input("> ").strip()
            if not line:
                return {}
            return json.loads(line)
        except (json.JSONDecodeError, EOFError) as e:
            print(f"Error: Invalid JSON: {e}")
            sys.exit(1)
`

const sdkToolsTemplate = `"""
Aphelion SDK — Tool execution
Calls the Aphelion gateway to execute subscribed tools.
"""

import os

import httpx


def _get_config():
    """Read runtime config from environment (injected by CLI)."""
    token = os.environ.get("APHELION_API_TOKEN", "")
    if not token:
        raise RuntimeError(
            "Agent not authenticated. "
            "Run: aphelion agent run agent.py (do not run agent.py directly)"
        )
    api_url = os.environ.get("APHELION_API_URL", "https://api.aphl.ai")
    session_id = os.environ.get("APHELION_SESSION_ID", "")
    return token, api_url, session_id


async def execute(tool_name: str, params: dict) -> dict:
    """Execute a tool via the Aphelion gateway.

    POST /v1/agents/{session_id}/execute

    Args:
        tool_name: Fully qualified tool name (e.g. "twilio.api20100401message.createmessage")
        params: Parameters to pass to the tool

    Returns:
        Response body as a dict

    Raises:
        RuntimeError: If not authenticated or the call fails
    """
    token, api_url, session_id = _get_config()

    url = f"{api_url}/v1/agents/{session_id}/execute"

    async with httpx.AsyncClient(timeout=60.0) as client:
        resp = await client.post(
            url,
            json={"tool": tool_name, "params": params},
            headers={
                "Authorization": f"Bearer {token}",
                "Content-Type": "application/json",
            },
        )

    if resp.status_code == 401:
        raise RuntimeError(
            "Session expired. Run: aphelion auth login"
        )
    if resp.status_code == 404:
        raise RuntimeError(
            f'Tool "{tool_name}" not found. '
            f"Check the name with: aphelion tools describe {tool_name}"
        )
    if resp.status_code >= 400:
        try:
            detail = resp.json().get("detail", resp.text)
        except Exception:
            detail = resp.text
        raise RuntimeError(
            f'Tool "{tool_name}" failed (HTTP {resp.status_code}).\n'
            f"Reason: {detail}\n"
            f"Docs:   aphelion tools describe {tool_name}"
        )

    data = resp.json()
    if data.get("success"):
        return data.get("result", data)
    return data


async def list() -> list:
    """List tools the current agent is subscribed to.

    Returns:
        List of tool descriptors
    """
    token, api_url, session_id = _get_config()

    url = f"{api_url}/v1/agents/{session_id}/tools"

    async with httpx.AsyncClient(timeout=30.0) as client:
        resp = await client.get(
            url,
            headers={"Authorization": f"Bearer {token}"},
        )

    if resp.status_code >= 400:
        raise RuntimeError(f"Failed to list tools (HTTP {resp.status_code}): {resp.text}")

    return resp.json()
`

const sdkMemoryTemplate = `"""
Aphelion SDK — Agent-scoped memory
Persistent key-value store that survives across executions.
"""

import os
from typing import Optional

import httpx


def _get_config():
    """Read runtime config from environment (injected by CLI)."""
    token = os.environ.get("APHELION_API_TOKEN", "")
    if not token:
        raise RuntimeError(
            "Agent not authenticated. "
            "Run: aphelion agent run agent.py (do not run agent.py directly)"
        )
    api_url = os.environ.get("APHELION_API_URL", "https://api.aphl.ai")
    agent_id = os.environ.get("APHELION_AGENT_ID", "")
    if not agent_id:
        raise RuntimeError(
            "APHELION_AGENT_ID not set. "
            "Run: aphelion agent run agent.py (do not run agent.py directly)"
        )
    return token, api_url, agent_id


async def get(key: str) -> Optional[dict]:
    """Read a memory entry by key.

    GET /v2/agents/{agent_id}/memory/{key}

    Returns:
        The stored value as a dict, or None if the key does not exist.
    """
    token, api_url, agent_id = _get_config()

    url = f"{api_url}/v2/agents/{agent_id}/memory/{key}"

    async with httpx.AsyncClient(timeout=30.0) as client:
        resp = await client.get(
            url,
            headers={"Authorization": f"Bearer {token}"},
        )

    if resp.status_code == 404:
        return None
    if resp.status_code >= 400:
        raise RuntimeError(f"Memory get failed (HTTP {resp.status_code}): {resp.text}")

    data = resp.json()
    if not data:
        return None
    # Unwrap {"value": ...} envelope if present
    if isinstance(data, dict) and "value" in data:
        return data["value"] or None
    return data


async def set(key: str, value: dict, ttl: str = None) -> None:
    """Write a memory entry.

    PUT /v2/agents/{agent_id}/memory/{key}

    Args:
        key: Memory key
        value: Dict to store
        ttl: Optional time-to-live (e.g. "7d", "24h", "30m")
    """
    token, api_url, agent_id = _get_config()

    url = f"{api_url}/v2/agents/{agent_id}/memory/{key}"
    body: dict = {"value": value}
    if ttl is not None:
        body["ttl"] = ttl

    async with httpx.AsyncClient(timeout=30.0) as client:
        resp = await client.put(
            url,
            json=body,
            headers={
                "Authorization": f"Bearer {token}",
                "Content-Type": "application/json",
            },
        )

    if resp.status_code >= 400:
        raise RuntimeError(f"Memory set failed (HTTP {resp.status_code}): {resp.text}")


async def delete(key: str) -> None:
    """Delete a memory entry.

    DELETE /v2/agents/{agent_id}/memory/{key}
    """
    token, api_url, agent_id = _get_config()

    url = f"{api_url}/v2/agents/{agent_id}/memory/{key}"

    async with httpx.AsyncClient(timeout=30.0) as client:
        resp = await client.delete(
            url,
            headers={"Authorization": f"Bearer {token}"},
        )

    if resp.status_code >= 400 and resp.status_code != 404:
        raise RuntimeError(f"Memory delete failed (HTTP {resp.status_code}): {resp.text}")


async def search(query: str, threshold: float = 0.7) -> list:
    """Semantic search across agent memory.

    POST /v2/agents/{agent_id}/memory/search

    Args:
        query: Natural language search query
        threshold: Similarity threshold (0.0 to 1.0, default 0.7)

    Returns:
        List of matching memory entries
    """
    token, api_url, agent_id = _get_config()

    url = f"{api_url}/v2/agents/{agent_id}/memory/search"

    async with httpx.AsyncClient(timeout=30.0) as client:
        resp = await client.post(
            url,
            json={"query": query, "threshold": threshold},
            headers={
                "Authorization": f"Bearer {token}",
                "Content-Type": "application/json",
            },
        )

    if resp.status_code >= 400:
        raise RuntimeError(f"Memory search failed (HTTP {resp.status_code}): {resp.text}")

    return resp.json()
`

const packageJSONTemplate = `{
  "name": "{{.Name}}",
  "version": "1.0.0",
  "description": "{{.Description}}",
  "main": "index.js",
  "scripts": {
    "start": "node index.js",
    "test": "node --test tests/test_agent.js"
  },
  "dependencies": {
    "@aphelion/sdk": "^0.1.0"
  }
}
`

const gitignoreTemplate = `config.yaml
.env
*.log
`

const projectConfigTemplate = `name: {{.Name}}
description: {{.Description}}
version: 1.0.0
language: {{.Language}}

agent:
  id: {{.AgentID}}
  client_id: {{.ClientID}}
  client_secret: {{.ClientSecret}}

gateway:
  api_url: https://api.aphl.ai

tools:
  subscribed:
{{.ToolsList}}

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
`

const readmeTemplate = `# {{.Name}}

{{.Description}}

## Quick Start

1. Install dependencies:

` + "```bash" + `
{{.InstallCmd}}
` + "```" + `

2. Copy environment variables and fill in your values:

` + "```bash" + `
cp .env.example .env
` + "```" + `

3. Run locally:

` + "```bash" + `
aphelion agent run {{.EntryFile}} --input '{"key": "value"}'
` + "```" + `

4. Deploy to Aphelion Cloud:

` + "```bash" + `
aphelion deploy
` + "```" + `

5. Invoke the deployed agent:

` + "```bash" + `
aphelion invoke {{.Name}} --input '{"key": "value"}'
` + "```" + `

## Project Structure

` + "```" + `
{{.Name}}/
├── {{.EntryFile}}          # Main agent logic
├── {{.DepsFile}}           # Dependencies
├── .aphelion/
│   ├── config.yaml         # Agent configuration (git-ignored)
│   ├── agent.json          # Input/output manifest
│   └── .gitignore
├── .env.example            # Environment variable template
├── README.md
└── tests/
    └── {{.TestFile}}       # Test scaffold
` + "```" + `

## Commands

| Command | Description |
|---------|-------------|
| ` + "`aphelion agent run {{.EntryFile}}`" + ` | Run agent locally |
| ` + "`aphelion deploy`" + ` | Deploy to Aphelion Cloud |
| ` + "`aphelion invoke {{.Name}}`" + ` | Invoke deployed agent |
| ` + "`aphelion status`" + ` | Show project status |
| ` + "`aphelion memory list`" + ` | List memory entries |
| ` + "`aphelion deployments logs {{.Name}}`" + ` | View deployment logs |
| ` + "`aphelion tools search \"<query>\"`" + ` | Search marketplace for tools |
| ` + "`aphelion tools subscribe <name>`" + ` | Subscribe to a tool |
`

const testPythonTemplate = `"""
Tests for {{.Name}}
Run with: python -m pytest tests/test_agent.py
"""

import pytest


class TestAgent:
    """Test suite for the agent."""

    def test_placeholder(self):
        """Replace this with real tests for your agent."""
        # Import your agent's main function and test it:
        #
        # from agent import main
        # import asyncio
        #
        # result = asyncio.run(main({"key": "value"}))
        # assert result["status"] == "success"
        pass
`

const testNodeTemplate = `/**
 * Tests for {{.Name}}
 * Run with: node --test tests/test_agent.js
 */

const { describe, it } = require("node:test");
const assert = require("node:assert");

describe("Agent", () => {
  it("placeholder — replace with real tests", () => {
    // Import your agent and test it:
    //
    // const agent = require("../index.js");
    // const result = await agent.run({ key: "value" });
    // assert.strictEqual(result.status, "success");
    assert.ok(true);
  });
});
`
