package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	authPkg "github.com/Exmplr-AI/aphelion-cli/pkg/auth"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
)

func newLoginCmd() *cobra.Command {
	var username, password string
	var noLaunchBrowser bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Aphelion Gateway",
		Long:  "Authenticate with your Aphelion Gateway account using OAuth2 or traditional credentials",
		Example: `  # Login with browser-based OAuth (recommended)
  aphelion auth login

  # Login without launching browser (copy URL manually)
  aphelion auth login --no-launch-browser

  # Login with username/password (legacy)
  aphelion auth login --username myuser

  # Login with both username and password flags (not recommended)
  aphelion auth login --username myuser --password mypass`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If username is provided, use legacy authentication
			if username != "" {
				return legacyLogin(username, password)
			}

			// Use OAuth flow
			return oauthLogin(noLaunchBrowser)
		},
	}

	cmd.Flags().StringVarP(&username, "username", "u", "", "username for legacy authentication")
	cmd.Flags().StringVarP(&password, "password", "p", "", "password for legacy authentication (not recommended, use interactive mode)")
	cmd.Flags().BoolVar(&noLaunchBrowser, "no-launch-browser", false, "print auth URL instead of opening browser")

	return cmd
}

func oauthLogin(noLaunchBrowser bool) error {
	client := api.NewClient()
	
	// Get OAuth configuration from API
	utils.PrintInfo("Getting authentication configuration...")
	var oauthConfig authPkg.OAuthConfig
	if err := client.Get("/auth/info", &oauthConfig); err != nil {
		return fmt.Errorf("failed to get OAuth configuration: %w", err)
	}

	// Set default client_id if not provided by API
	if oauthConfig.ClientID == "" {
		oauthConfig.ClientID = "UbXxpQBSr9AsqpS2ln2jzmsmamromaFC" // Correct client_id
	}

	// Update redirect URI to use localhost
	oauthConfig.RedirectURI = "http://localhost:8765/callback"
	
	if noLaunchBrowser {
		// Generate PKCE for manual mode
		codeVerifier, codeChallenge, err := authPkg.GeneratePKCE()
		if err != nil {
			return fmt.Errorf("failed to generate PKCE parameters: %w", err)
		}
		oauthConfig.CodeVerifier = codeVerifier
		oauthConfig.CodeChallenge = codeChallenge

		// Generate state parameter for CSRF protection
		state, err := authPkg.GenerateState()
		if err != nil {
			return fmt.Errorf("failed to generate state parameter: %w", err)
		}
		oauthConfig.State = state

		// Build auth URL with PKCE (includes offline_access for refresh tokens)
		// prompt=login forces a fresh Auth0 session, avoiding "Invalid session" errors
		authURL := fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&scope=openid%%20profile%%20email%%20offline_access&audience=%s&code_challenge=%s&code_challenge_method=S256&state=%s&prompt=login",
			oauthConfig.AuthURL,
			oauthConfig.ClientID,
			oauthConfig.RedirectURI,
			oauthConfig.Audience,
			oauthConfig.CodeChallenge,
			oauthConfig.State,
		)
		
		fmt.Printf("\nPlease open the following URL in your browser:\n\n%s\n\n", authURL)
		fmt.Print("Enter the authorization code: ")
		
		var code string
		if _, err := fmt.Scanln(&code); err != nil {
			return fmt.Errorf("failed to read authorization code: %w", err)
		}
		
		return completeOAuthFlow(&oauthConfig, code)
	}

	// Start OAuth flow with browser
	result, err := authPkg.StartOAuthFlow(&oauthConfig)
	if err != nil {
		return fmt.Errorf("OAuth flow failed: %w", err)
	}

	if result.Error != "" {
		return fmt.Errorf("authentication failed: %s", result.Error)
	}

	return completeOAuthFlow(&oauthConfig, result.Code)
}

func completeOAuthFlow(oauthConfig *authPkg.OAuthConfig, code string) error {
	spinner := utils.NewSpinner("Exchanging authorization code for token...")
	spinner.Start()

	// Exchange code for token
	tokenResp, err := authPkg.ExchangeCodeForToken(oauthConfig, code)
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Create/get user in Aphelion database
	userInfo, err := authPkg.CreateOrGetUser(config.GetAPIUrl(), tokenResp.AccessToken)
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to create/get user profile: %w", err)
	}

	spinner.Stop()

	// Save authentication
	if err := config.SetAuth(tokenResp.AccessToken, tokenResp.RefreshToken, userInfo.Sub, userInfo.Email, userInfo.Name, tokenResp.ExpiresIn); err != nil {
		return fmt.Errorf("failed to save authentication: %w", err)
	}

	utils.PrintSuccess("Successfully authenticated as %s (%s)", userInfo.Name, userInfo.Email)
	return nil
}

func legacyLogin(username, password string) error {
	client := api.NewClient()

	if username == "" {
		fmt.Print("Username: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read username: %w", err)
		}
		username = strings.TrimSpace(input)
	}

	if password == "" {
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		password = string(passwordBytes)
		fmt.Println()
	}

	if username == "" || password == "" {
		return fmt.Errorf("username and password are required")
	}

	spinner := utils.NewSpinner("Authenticating...")
	spinner.Start()

	loginReq := api.LoginRequest{
		Username: username,
		Password: password,
	}

	var authResp api.AuthResponse
	err := client.Post("/auth/login", loginReq, &authResp)
	spinner.Stop()

	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if err := config.SetAuth(authResp.Token, "", authResp.User.ID, authResp.User.Email, authResp.User.Username, 0); err != nil {
		return fmt.Errorf("failed to save authentication: %w", err)
	}

	utils.PrintSuccess("Successfully logged in as %s", authResp.User.Username)
	return nil
}