package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/seth4242/snet/internal/buildinfo"
)

const (
	configDir  = ".snet"
	configFile = "config.json"
)

// DefaultAPIBase returns the default API base URL based on build mode
func DefaultAPIBase() string {
	return buildinfo.GetAPIBase()
}

// Config holds the CLI configuration
type Config struct {
	APIToken        string `json:"api_token"`
	AccountID       string `json:"account_id"`
	APIBase         string `json:"api_base"`
	DefaultProvider string `json:"default_provider,omitempty"` // "frp" or "cloudflare", defaults to "frp"
	DefaultWildcard *bool  `json:"default_wildcard,omitempty"` // Enable wildcard by default, defaults to true
}

// configPath returns the full path to the config file
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, configDir, configFile), nil
}

// Load reads the config from disk
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not logged in. Run 'snet login' first")
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set default API base if not specified
	if cfg.APIBase == "" {
		cfg.APIBase = DefaultAPIBase()
	}

	// Set default provider if not specified
	if cfg.DefaultProvider == "" {
		cfg.DefaultProvider = "frp"
	}

	return &cfg, nil
}

// WildcardDefault returns true if wildcard should be enabled by default
// Returns true if not explicitly set to false
func (c *Config) WildcardDefault() bool {
	if c.DefaultWildcard == nil {
		return true // Default to wildcard enabled
	}
	return *c.DefaultWildcard
}

// Save writes the config to disk
func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set default API base
	if cfg.APIBase == "" {
		cfg.APIBase = DefaultAPIBase()
	}

	// Set default provider
	if cfg.DefaultProvider == "" {
		cfg.DefaultProvider = "frp"
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Exists checks if a config file exists
func Exists() bool {
	path, err := configPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// Delete removes the config file
func Delete() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	return os.Remove(path)
}
