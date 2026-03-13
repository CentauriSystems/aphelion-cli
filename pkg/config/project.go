package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ProjectConfig struct {
	Name        string           `yaml:"name"`
	Description string           `yaml:"description"`
	Version     string           `yaml:"version"`
	Language    string           `yaml:"language"`
	Agent       AgentIdentity    `yaml:"agent,omitempty"`
	Gateway     GatewayConfig    `yaml:"gateway,omitempty"`
	Tools       ToolsConfig      `yaml:"tools,omitempty"`
	Execution   ExecutionConfig  `yaml:"execution,omitempty"`
	Deployment  DeploymentConfig `yaml:"deployment,omitempty"`
	Schedule    ScheduleConfig   `yaml:"schedule,omitempty"`
	Logging     LoggingConfig    `yaml:"logging,omitempty"`
}

type AgentIdentity struct {
	ID           string `yaml:"id,omitempty"`
	ClientID     string `yaml:"client_id,omitempty"`
	ClientSecret string `yaml:"client_secret,omitempty"`
}

type GatewayConfig struct {
	APIUrl string `yaml:"api_url,omitempty"`
}

type ToolsConfig struct {
	Subscribed []string `yaml:"subscribed,omitempty"`
}

type ExecutionConfig struct {
	Timeout        string `yaml:"timeout,omitempty"`
	MemoryAutoSave bool   `yaml:"memory_auto_save,omitempty"`
	MaxRetries     int    `yaml:"max_retries,omitempty"`
}

type DeploymentConfig struct {
	Status       string `yaml:"status,omitempty"`
	Endpoint     string `yaml:"endpoint,omitempty"`
	Region       string `yaml:"region,omitempty"`
	LastDeployed string `yaml:"last_deployed,omitempty"`
}

type ScheduleConfig struct {
	Cron    string `yaml:"cron,omitempty"`
	Enabled bool   `yaml:"enabled,omitempty"`
}

type LoggingConfig struct {
	Level string `yaml:"level,omitempty"`
	File  string `yaml:"file,omitempty"`
}

// IsProjectDir returns true if the current directory contains an Aphelion project.
func IsProjectDir() bool {
	_, err := os.Stat(filepath.Join(".aphelion", "config.yaml"))
	return err == nil
}

// LoadProjectConfig reads the project config from .aphelion/config.yaml in the current directory.
// Returns nil, nil if not in a project directory.
func LoadProjectConfig() (*ProjectConfig, error) {
	configPath := filepath.Join(".aphelion", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read project config: %w", err)
	}

	var cfg ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse project config: %w", err)
	}

	return &cfg, nil
}

// SaveProjectConfig writes the project config to .aphelion/config.yaml.
func SaveProjectConfig(cfg *ProjectConfig) error {
	configDir := ".aphelion"
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create .aphelion directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal project config: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write project config: %w", err)
	}

	return nil
}

// GetAgentID returns the agent ID from the project config, or empty string if not in a project.
func GetAgentID() string {
	cfg, err := LoadProjectConfig()
	if err != nil || cfg == nil {
		return ""
	}
	return cfg.Agent.ID
}

// GetAgentCredentials returns the client_id and client_secret from the project config.
func GetAgentCredentials() (clientID, clientSecret string) {
	cfg, err := LoadProjectConfig()
	if err != nil || cfg == nil {
		return "", ""
	}
	return cfg.Agent.ClientID, cfg.Agent.ClientSecret
}
