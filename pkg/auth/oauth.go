package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
)

const (
	callbackPort = "8765"
	callbackPath = "/callback"
	callbackURL  = "http://localhost:" + callbackPort + callbackPath
)

type OAuthConfig struct {
	Domain         string `json:"auth0_domain"`
	ClientID       string `json:"client_id"`
	Audience       string `json:"auth0_audience"`
	RedirectURI    string `json:"redirect_uri"`
	AuthURL        string `json:"auth_url"`
	CodeVerifier   string `json:"-"`
	CodeChallenge  string `json:"-"`
	State          string `json:"-"`
}

type AuthResult struct {
	Code  string
	Error string
}

// GeneratePKCE generates code verifier and challenge for PKCE
func GeneratePKCE() (string, string, error) {
	// Generate code verifier (43-128 characters)
	codeVerifier := make([]byte, 32)
	if _, err := rand.Read(codeVerifier); err != nil {
		return "", "", err
	}
	verifier := base64.RawURLEncoding.EncodeToString(codeVerifier)

	// Generate code challenge (SHA256 hash of verifier)
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return verifier, challenge, nil
}

// GenerateState generates a cryptographic random state parameter for CSRF protection
func GenerateState() (string, error) {
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(stateBytes), nil
}

// StartOAuthFlow starts the OAuth flow and returns the authorization code
func StartOAuthFlow(config *OAuthConfig) (*AuthResult, error) {
	// Generate PKCE parameters
	codeVerifier, codeChallenge, err := GeneratePKCE()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PKCE parameters: %w", err)
	}

	config.CodeVerifier = codeVerifier
	config.CodeChallenge = codeChallenge

	// Generate state parameter for CSRF protection
	state, err := GenerateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state parameter: %w", err)
	}
	config.State = state

	// Create authorization URL with PKCE (includes offline_access for refresh tokens)
	// prompt=login forces a fresh Auth0 session, avoiding "Invalid session" errors from stale cookies
	authURL := fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&scope=openid%%20profile%%20email%%20offline_access&audience=%s&code_challenge=%s&code_challenge_method=S256&state=%s&prompt=login",
		config.AuthURL,
		url.QueryEscape(config.ClientID),
		url.QueryEscape(callbackURL),
		url.QueryEscape(config.Audience),
		url.QueryEscape(codeChallenge),
		url.QueryEscape(state),
	)

	// Start local server to handle callback
	resultChan := make(chan *AuthResult, 1)
	server := &http.Server{
		Addr:    ":" + callbackPort,
		Handler: createCallbackHandler(resultChan, state),
	}

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			resultChan <- &AuthResult{Error: fmt.Sprintf("Failed to start callback server: %v", err)}
		}
	}()

	// Wait a moment for server to start
	time.Sleep(100 * time.Millisecond)

	// Open browser
	utils.OpenBrowserWithFallback(authURL)
	utils.PrintInfo("Waiting for authentication in browser...")
	utils.PrintInfo("You can close the browser tab after authentication completes.")

	// Wait for callback with timeout
	var result *AuthResult
	select {
	case result = <-resultChan:
	case <-time.After(5 * time.Minute):
		result = &AuthResult{Error: "Authentication timed out after 5 minutes"}
	}

	// Shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	return result, nil
}

func createCallbackHandler(resultChan chan<- *AuthResult, expectedState string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		query := r.URL.Query()
		code := query.Get("code")
		stateParam := query.Get("state")
		errorParam := query.Get("error")
		errorDescription := query.Get("error_description")

		// Validate state parameter (CSRF protection)
		if stateParam != expectedState {
			result := &AuthResult{Error: "Invalid state parameter — possible CSRF attack. Please try again."}
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Authentication Error - Aphelion CLI</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; margin-top: 50px; color: #333; }
        .error { color: #d32f2f; margin: 20px; }
        .container { max-width: 500px; margin: 0 auto; padding: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Authentication Error</h1>
        <div class="error">Invalid state parameter. Please try again.</div>
        <p>Please return to your terminal and try again.</p>
        <p>You can close this tab.</p>
    </div>
</body>
</html>`)
			select {
			case resultChan <- result:
			default:
			}
			return
		}

		var result *AuthResult

		if errorParam != "" {
			errorMsg := errorParam
			if errorDescription != "" {
				errorMsg += ": " + errorDescription
			}
			result = &AuthResult{Error: errorMsg}
			
			// Send error response to browser
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Authentication Error - Aphelion CLI</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; margin-top: 50px; color: #333; }
        .error { color: #d32f2f; margin: 20px; }
        .container { max-width: 500px; margin: 0 auto; padding: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Authentication Error</h1>
        <div class="error">%s</div>
        <p>Please return to your terminal and try again.</p>
        <p>You can close this tab.</p>
    </div>
</body>
</html>`, errorMsg)
		} else if code != "" {
			result = &AuthResult{Code: code}
			
			// Redirect to success page
			w.Header().Set("Location", "https://console.aphl.ai/auth/success")
			w.WriteHeader(http.StatusFound)
		} else {
			result = &AuthResult{Error: "No authorization code received"}
			
			// Send error response
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Authentication Error - Aphelion CLI</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; margin-top: 50px; color: #333; }
        .error { color: #d32f2f; margin: 20px; }
        .container { max-width: 500px; margin: 0 auto; padding: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Authentication Error</h1>
        <div class="error">No authorization code received</div>
        <p>Please return to your terminal and try again.</p>
        <p>You can close this tab.</p>
    </div>
</body>
</html>`)
		}

		// Send result to channel
		select {
		case resultChan <- result:
		default:
		}
	}
}