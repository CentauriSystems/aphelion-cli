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

// StartOAuthFlow starts the OAuth flow and returns the authorization code
func StartOAuthFlow(config *OAuthConfig) (*AuthResult, error) {
	// Generate PKCE parameters
	codeVerifier, codeChallenge, err := GeneratePKCE()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PKCE parameters: %w", err)
	}
	
	config.CodeVerifier = codeVerifier
	config.CodeChallenge = codeChallenge

	// Create authorization URL with PKCE
	authURL := fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&scope=openid%%20profile%%20email&audience=%s&code_challenge=%s&code_challenge_method=S256",
		config.AuthURL,
		url.QueryEscape(config.ClientID),
		url.QueryEscape(callbackURL),
		url.QueryEscape(config.Audience),
		url.QueryEscape(codeChallenge),
	)

	// Start local server to handle callback
	resultChan := make(chan *AuthResult, 1)
	server := &http.Server{
		Addr:    ":" + callbackPort,
		Handler: createCallbackHandler(resultChan),
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

func createCallbackHandler(resultChan chan<- *AuthResult) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		query := r.URL.Query()
		code := query.Get("code")
		errorParam := query.Get("error")
		errorDescription := query.Get("error_description")

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