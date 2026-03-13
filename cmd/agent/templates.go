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
  const patientName = input.patient_name;
  const contact = input.contact;
  const visitType = input.visit_type || "your recent visit";

  if (!patientName || !contact) {
    return { status: "failed", error: "patient_name and contact are required" };
  }

  // Check if we've contacted this patient recently (avoid duplicate requests)
  const recent = await memory.get(` + "`last_request:${contact}`" + `);
  if (recent && (recent.sent_within_days || 0) < 7) {
    agent.log(` + "`Skipping ${contact} — already contacted within 7 days`" + `);
    return { status: "skipped", reason: "recently_contacted" };
  }

  const reviewLink = agent.env("REVIEW_LINK", "https://g.page/r/your-review-link");
  const message =
    ` + "`Hi ${patientName}, thank you for ${visitType}. " +
		"We'd love your feedback — it only takes 30 seconds: ${reviewLink}`" + `;

  const isPhone = contact.startsWith("+") || /^[\\d\\s-]+$/.test(contact);

  let result;
  let channel;

  if (isPhone) {
    result = await tools.execute("twilio.messaging.create_message", {
      To: contact,
      Body: message,
      From: agent.env("TWILIO_PHONE_NUMBER"),
    });
    channel = "sms";
  } else {
    result = await tools.execute("sendgrid.mail.send", {
      personalizations: [{ to: [{ email: contact }] }],
      from: { email: agent.env("SENDGRID_FROM_EMAIL") },
      subject: ` + "`How was your experience, ${patientName}?`" + `,
      content: [{ type: "text/plain", value: message }],
    });
    channel = "email";
  }

  // Write to memory — persists across executions, scoped to this agent
  await memory.set(
    ` + "`last_request:${contact}`" + `,
    {
      patient: patientName,
      channel,
      visit_type: visitType,
      sent_within_days: 0,
    },
    { ttl: "7d" }
  );

  return {
    status: result.success ? "sent" : "failed",
    channel,
    timestamp: result.timestamp || null,
    error: result.error || null,
  };
});

if (require.main === module) {
  // Run locally: node index.js
  // Or:          aphelion agent run index.js --input '{"patient_name": "Jane", "contact": "+15551234567"}'
  agent.runLocal();
}
`

const agentJSONTemplate = `{
  "name": "{{.Name}}",
  "description": "{{.Description}}",
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
`

const envExampleTemplate = `# Copy to .env and fill in your values
# .env is git-ignored — never commit secrets

TWILIO_PHONE_NUMBER=+1XXXXXXXXXX
SENDGRID_FROM_EMAIL=noreply@yourdomain.com
REVIEW_LINK=https://g.page/r/your-review-link
`

const requirementsTemplate = `# Aphelion SDK — provides Agent, tools, memory imports
aphelion-sdk>=0.1.0

# Add your own dependencies below
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
aphelion agent run {{.EntryFile}} --input '{"patient_name": "Jane", "contact": "+15551234567"}'
` + "```" + `

4. Deploy to Aphelion Cloud:

` + "```bash" + `
aphelion deploy
` + "```" + `

5. Invoke the deployed agent:

` + "```bash" + `
aphelion invoke {{.Name}} --input '{"patient_name": "Jane", "contact": "+15551234567"}'
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

## Environment Variables

| Variable | Description |
|----------|-------------|
| ` + "`TWILIO_PHONE_NUMBER`" + ` | Your Twilio phone number |
| ` + "`SENDGRID_FROM_EMAIL`" + ` | Sender email for SendGrid |
| ` + "`REVIEW_LINK`" + ` | Link to your review page |
`

const testPythonTemplate = `"""
Tests for {{.Name}}
Run with: python -m pytest tests/test_agent.py
"""

import pytest


class TestAgent:
    """Test suite for the agent."""

    def test_missing_patient_name(self):
        """Agent should fail if patient_name is missing."""
        input_data = {"contact": "+15551234567"}
        # TODO: Import and call your agent's main function
        # result = await main(input_data)
        # assert result["status"] == "failed"
        # assert "patient_name" in result["error"]
        pass

    def test_missing_contact(self):
        """Agent should fail if contact is missing."""
        input_data = {"patient_name": "Jane"}
        # TODO: Import and call your agent's main function
        # result = await main(input_data)
        # assert result["status"] == "failed"
        # assert "contact" in result["error"]
        pass

    def test_phone_detection(self):
        """Agent should detect phone numbers starting with +."""
        # Contacts starting with "+" or all digits should route to SMS
        assert "+15551234567".startswith("+")
        assert "5551234567".replace("-", "").replace(" ", "").isdigit()

    def test_email_detection(self):
        """Agent should detect email addresses."""
        contact = "jane@example.com"
        is_phone = contact.startswith("+") or contact.replace("-", "").replace(" ", "").isdigit()
        assert not is_phone

    def test_valid_input(self):
        """Agent should process valid input successfully."""
        input_data = {
            "patient_name": "Jane Smith",
            "contact": "+15551234567",
            "visit_type": "dental cleaning",
        }
        # TODO: Mock tools.execute and memory calls, then test
        # result = await main(input_data)
        # assert result["status"] == "sent"
        # assert result["channel"] == "sms"
        pass
`

const testNodeTemplate = `/**
 * Tests for {{.Name}}
 * Run with: node --test tests/test_agent.js
 */

const { describe, it } = require("node:test");
const assert = require("node:assert");

describe("Agent", () => {
  it("should detect phone numbers", () => {
    const contact = "+15551234567";
    const isPhone = contact.startsWith("+") || /^[\d\s-]+$/.test(contact);
    assert.strictEqual(isPhone, true);
  });

  it("should detect email addresses", () => {
    const contact = "jane@example.com";
    const isPhone = contact.startsWith("+") || /^[\d\s-]+$/.test(contact);
    assert.strictEqual(isPhone, false);
  });

  it("should reject missing patient_name", () => {
    const input = { contact: "+15551234567" };
    assert.strictEqual(input.patient_name, undefined);
    // TODO: Import agent and test with mocked tools/memory
  });

  it("should reject missing contact", () => {
    const input = { patient_name: "Jane" };
    assert.strictEqual(input.contact, undefined);
    // TODO: Import agent and test with mocked tools/memory
  });
});
`
