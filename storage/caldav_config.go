package storage

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

// CalDAVConfig holds configuration for CalDAV server connection
type CalDAVConfig struct {
	Enabled         bool   `json:"enabled"`
	ServerURL       string `json:"server_url"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	CalendarHomeURL string `json:"calendar_home_url"` // Full URL to calendar home (will discover all calendars under this)
	SyncInterval    int    `json:"sync_interval"`     // Seconds between auto-sync (0 = manual only)
}

// DefaultCalDAVConfig returns a disabled default configuration
func DefaultCalDAVConfig() *CalDAVConfig {
	return &CalDAVConfig{
		Enabled:         false,
		ServerURL:       "",
		Username:        "",
		Password:        "",
		CalendarHomeURL: "",
		SyncInterval:    300, // 5 minutes default for auto-sync
	}
}

// Validate checks if the CalDAV configuration is valid
func (c *CalDAVConfig) Validate() error {
	if !c.Enabled {
		return nil // No validation needed if disabled
	}

	if c.ServerURL == "" {
		return fmt.Errorf("server URL is required")
	}

	if c.Username == "" {
		return fmt.Errorf("username is required")
	}

	if c.Password == "" {
		return fmt.Errorf("password is required")
	}

	// Enforce HTTPS for security (Basic Auth over HTTP is insecure)
	parsedURL, err := url.Parse(c.ServerURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	if parsedURL.Scheme != "https" {
		// Only allow HTTP for actual localhost addresses
		hostname := parsedURL.Hostname()
		isLocalhost := hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1"

		if !isLocalhost {
			return fmt.Errorf("server URL must use HTTPS for security (HTTP only allowed for localhost/127.0.0.1/::1)")
		}
	}

	return nil
}

// GetConfigPath returns the path to the CalDAV config file
// Always uses ~/.config/chronos/caldav_config.json for consistency across platforms
func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.Getenv("HOME")
	}
	return filepath.Join(homeDir, ".config", "chronos", "caldav_config.json")
}

// LoadCalDAVConfig loads CalDAV configuration from disk
func LoadCalDAVConfig() (*CalDAVConfig, error) {
	configPath := GetConfigPath()

	// If config doesn't exist, return default (disabled)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultCalDAVConfig(), nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Parse JSON
	var config CalDAVConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// SaveCalDAVConfig saves CalDAV configuration to disk
func SaveCalDAVConfig(config *CalDAVConfig) error {
	configPath := GetConfigPath()

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0o600); err != nil { // 0600 for security (credentials)
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
