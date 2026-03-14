package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// AuthConfig holds nested auth fields in the new config format.
type AuthConfig struct {
	AccessToken  string `yaml:"access_token,omitempty" mapstructure:"access_token"`
	RefreshToken string `yaml:"refresh_token,omitempty" mapstructure:"refresh_token"`
	ExpiresAt    string `yaml:"expires_at,omitempty" mapstructure:"expires_at"`
	UserEmail    string `yaml:"user_email,omitempty" mapstructure:"user_email"`
	AccountID    string `yaml:"account_id,omitempty" mapstructure:"account_id"`
	Username     string `yaml:"username,omitempty" mapstructure:"username"`
}

// Config is the top-level config structure written to ~/.aphelion/config.yaml.
type Config struct {
	APIUrl string     `yaml:"api_url" mapstructure:"api_url"`
	Output string     `yaml:"output" mapstructure:"output"`
	Auth   AuthConfig `yaml:"auth,omitempty" mapstructure:"auth"`

	// Legacy fields — read for backward compat, not written.
	LegacyAccessToken string    `yaml:"access_token,omitempty" mapstructure:"access_token"`
	LegacyUserID      string    `yaml:"user_id,omitempty" mapstructure:"user_id"`
	LegacyEmail       string    `yaml:"email,omitempty" mapstructure:"email"`
	LegacyUsername    string    `yaml:"username,omitempty" mapstructure:"username"`
	LegacyLastLogin   time.Time `yaml:"last_login,omitempty" mapstructure:"last_login"`
}

// configForSave is the struct used when writing config to disk.
// It omits legacy fields so only the new nested auth structure is persisted.
type configForSave struct {
	APIUrl string     `yaml:"api_url"`
	Output string     `yaml:"output"`
	Auth   AuthConfig `yaml:"auth,omitempty"`
}

var globalConfig *Config

// RefreshFunc is a callback for refreshing an expired access token.
// It is set externally (e.g. in cmd/root.go) to avoid circular imports.
var RefreshFunc func(refreshToken string) (newAccessToken string, newExpiresIn int, err error)

func InitConfig() {
	viper.SetDefault("api_url", "https://api.aphl.ai")
	viper.SetDefault("output", "table")

	globalConfig = &Config{
		APIUrl: viper.GetString("api_url"),
		Output: viper.GetString("output"),
	}

	// Read new nested auth fields
	if viper.IsSet("auth.access_token") {
		globalConfig.Auth.AccessToken = viper.GetString("auth.access_token")
	}
	if viper.IsSet("auth.refresh_token") {
		globalConfig.Auth.RefreshToken = viper.GetString("auth.refresh_token")
	}
	if viper.IsSet("auth.expires_at") {
		globalConfig.Auth.ExpiresAt = viper.GetString("auth.expires_at")
	}
	if viper.IsSet("auth.user_email") {
		globalConfig.Auth.UserEmail = viper.GetString("auth.user_email")
	}
	if viper.IsSet("auth.account_id") {
		globalConfig.Auth.AccountID = viper.GetString("auth.account_id")
	}
	if viper.IsSet("auth.username") {
		globalConfig.Auth.Username = viper.GetString("auth.username")
	}

	// Read legacy top-level fields
	if viper.IsSet("access_token") {
		globalConfig.LegacyAccessToken = viper.GetString("access_token")
	}
	if viper.IsSet("user_id") {
		globalConfig.LegacyUserID = viper.GetString("user_id")
	}
	if viper.IsSet("email") {
		globalConfig.LegacyEmail = viper.GetString("email")
	}
	if viper.IsSet("username") {
		globalConfig.LegacyUsername = viper.GetString("username")
	}
	if viper.IsSet("last_login") {
		globalConfig.LegacyLastLogin = viper.GetTime("last_login")
	}

	// Migrate legacy fields into Auth if Auth is empty
	if globalConfig.Auth.AccessToken == "" && globalConfig.LegacyAccessToken != "" {
		globalConfig.Auth.AccessToken = globalConfig.LegacyAccessToken
	}
	if globalConfig.Auth.UserEmail == "" && globalConfig.LegacyEmail != "" {
		globalConfig.Auth.UserEmail = globalConfig.LegacyEmail
	}
	if globalConfig.Auth.AccountID == "" && globalConfig.LegacyUserID != "" {
		globalConfig.Auth.AccountID = globalConfig.LegacyUserID
	}
	if globalConfig.Auth.Username == "" && globalConfig.LegacyUsername != "" {
		globalConfig.Auth.Username = globalConfig.LegacyUsername
	}
}

func GetConfig() *Config {
	if globalConfig == nil {
		InitConfig()
	}
	return globalConfig
}

func SaveConfig() error {
	if globalConfig == nil {
		return fmt.Errorf("config not initialized")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".aphelion")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")

	// Save only the new format (no legacy fields)
	toSave := configForSave{
		APIUrl: globalConfig.APIUrl,
		Output: globalConfig.Output,
		Auth:   globalConfig.Auth,
	}

	data, err := yaml.Marshal(toSave)
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return fmt.Errorf("could not write config file: %w", err)
	}

	// Update viper with new nested keys
	viper.Set("auth.access_token", globalConfig.Auth.AccessToken)
	viper.Set("auth.refresh_token", globalConfig.Auth.RefreshToken)
	viper.Set("auth.expires_at", globalConfig.Auth.ExpiresAt)
	viper.Set("auth.user_email", globalConfig.Auth.UserEmail)
	viper.Set("auth.account_id", globalConfig.Auth.AccountID)
	viper.Set("auth.username", globalConfig.Auth.Username)

	return nil
}

// SetAuth stores authentication tokens and user info.
func SetAuth(token, refreshToken, userID, email, username string, expiresIn int) error {
	config := GetConfig()
	// Ensure API URL is always correct
	config.APIUrl = "https://api.aphl.ai"
	config.Auth.AccessToken = token
	config.Auth.RefreshToken = refreshToken
	config.Auth.AccountID = userID
	config.Auth.UserEmail = email
	config.Auth.Username = username
	if expiresIn > 0 {
		config.Auth.ExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second).Format(time.RFC3339)
	}

	return SaveConfig()
}

// ClearAuth removes all stored authentication data.
func ClearAuth() error {
	config := GetConfig()
	config.Auth = AuthConfig{}
	// Clear legacy fields too
	config.LegacyAccessToken = ""
	config.LegacyUserID = ""
	config.LegacyEmail = ""
	config.LegacyUsername = ""
	config.LegacyLastLogin = time.Time{}

	return SaveConfig()
}

// IsAuthenticated returns true if a valid access token is available.
func IsAuthenticated() bool {
	return GetAccessToken() != ""
}

func GetAPIUrl() string {
	return GetConfig().APIUrl
}

// GetAccessToken returns a valid access token, refreshing if necessary.
func GetAccessToken() string {
	cfg := GetConfig()

	token := cfg.Auth.AccessToken
	if token == "" {
		// Fall back to legacy
		token = cfg.LegacyAccessToken
	}
	if token == "" {
		return ""
	}

	// Check expiry
	if cfg.Auth.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, cfg.Auth.ExpiresAt)
		if err == nil && time.Now().After(expiresAt) {
			// Token is expired — try to refresh
			if cfg.Auth.RefreshToken != "" && RefreshFunc != nil {
				newToken, expiresIn, err := RefreshFunc(cfg.Auth.RefreshToken)
				if err == nil && newToken != "" {
					cfg.Auth.AccessToken = newToken
					cfg.Auth.ExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second).Format(time.RFC3339)
					_ = SaveConfig()
					return newToken
				}
			}
			// Refresh failed or unavailable
			return ""
		}
	}

	return token
}

// GetUserEmail returns the authenticated user's email.
func GetUserEmail() string {
	cfg := GetConfig()
	if cfg.Auth.UserEmail != "" {
		return cfg.Auth.UserEmail
	}
	return cfg.LegacyEmail
}

// GetAccountID returns the authenticated user's account/user ID.
func GetAccountID() string {
	cfg := GetConfig()
	if cfg.Auth.AccountID != "" {
		return cfg.Auth.AccountID
	}
	return cfg.LegacyUserID
}

// GetUsername returns the authenticated user's username.
func GetUsername() string {
	cfg := GetConfig()
	if cfg.Auth.Username != "" {
		return cfg.Auth.Username
	}
	return cfg.LegacyUsername
}

// GetUserID is a backward-compatible alias for GetAccountID.
func GetUserID() string {
	return GetAccountID()
}

func GetOutputFormat() string {
	return GetConfig().Output
}
