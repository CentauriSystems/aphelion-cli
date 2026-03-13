package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newQuickstartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quickstart",
		Short: "Interactive quickstart guide for Aphelion",
		Long:  "Print a step-by-step guide to get started with Aphelion: authenticate, create an agent, deploy, and invoke.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(`
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Aphelion Quickstart Guide
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Step 1: Authenticate
  aphelion auth login

Step 2: Initialize an agent project
  aphelion agent init

Step 3: Subscribe to tools
  aphelion tools subscribe twilio
  aphelion tools subscribe sendgrid

Step 4: Set up environment variables
  cp .env.example .env
  # Edit .env with your API keys

Step 5: Run your agent locally
  aphelion agent run agent.py --input '{"patient_name": "Jane", "contact": "+15551234567"}'

Step 6: Deploy to Aphelion Cloud
  aphelion deploy

Step 7: Set environment variables for deployment
  aphelion env set TWILIO_PHONE_NUMBER "+15551234567"
  aphelion env set SENDGRID_FROM_EMAIL "noreply@yourdomain.com"

Step 8: Invoke your deployed agent
  aphelion invoke <agent-name> --input '{"patient_name": "Jane", "contact": "+15551234567"}'

Step 9: Check status
  aphelion status

Step 10: Set up for Claude (optional)
  aphelion mcp config

For full documentation: https://api.aphl.ai/docs
`)
		},
	}

	return cmd
}
