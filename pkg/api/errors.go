package api

import (
	"fmt"
	"strings"
)

// ActionableError wraps an API error with a user-friendly message and suggested action.
type ActionableError struct {
	StatusCode int
	Message    string
	Suggestion string
}

func (e *ActionableError) Error() string {
	if e.Suggestion != "" {
		return fmt.Sprintf("%s\n%s", e.Message, e.Suggestion)
	}
	return e.Message
}

// toActionableError converts a raw API error into a user-friendly error with suggestions.
func toActionableError(statusCode int, endpoint string, rawMessage string) error {
	switch statusCode {
	case 401:
		return &ActionableError{
			StatusCode: 401,
			Message:    "Session expired.",
			Suggestion: "Run: aphelion auth login",
		}
	case 403:
		return &ActionableError{
			StatusCode: 403,
			Message:    "Permission denied.",
			Suggestion: "Check your account permissions or contact your administrator.",
		}
	case 404:
		return notFoundError(endpoint, rawMessage)
	case 409:
		return &ActionableError{
			StatusCode: 409,
			Message:    fmt.Sprintf("Conflict: %s", rawMessage),
			Suggestion: "The resource may already exist. Check with: aphelion agents list",
		}
	case 422:
		return &ActionableError{
			StatusCode: 422,
			Message:    fmt.Sprintf("Validation error: %s", rawMessage),
		}
	case 429:
		return &ActionableError{
			StatusCode: 429,
			Message:    "Rate limit exceeded.",
			Suggestion: "Wait a moment and try again.",
		}
	case 500, 502, 503:
		return &ActionableError{
			StatusCode: statusCode,
			Message:    "The Aphelion API is experiencing issues.",
			Suggestion: "Check status at https://beta.console.aphl.ai or try again shortly.",
		}
	}

	if rawMessage != "" {
		return &ActionableError{StatusCode: statusCode, Message: rawMessage}
	}
	return &ActionableError{StatusCode: statusCode, Message: fmt.Sprintf("API error (HTTP %d)", statusCode)}
}

func notFoundError(endpoint string, rawMessage string) error {
	e := &ActionableError{StatusCode: 404}

	switch {
	case strings.Contains(endpoint, "/agents/"):
		e.Message = "Agent not found."
		e.Suggestion = "List your agents: aphelion agents list"
	case strings.Contains(endpoint, "/memory/"):
		e.Message = "Memory key not found."
	case strings.Contains(endpoint, "/tools/"):
		e.Message = "Tool not found."
		e.Suggestion = "Search available tools: aphelion tools search <query>"
	case strings.Contains(endpoint, "/deployments/") || strings.Contains(endpoint, "/deploy"):
		e.Message = "Deployment not found."
		e.Suggestion = "List deployments: aphelion deployments list"
	case strings.Contains(endpoint, "/services/"):
		e.Message = "Service not found."
		e.Suggestion = "List services: aphelion registry list"
	default:
		e.Message = fmt.Sprintf("Not found: %s", rawMessage)
	}

	return e
}
