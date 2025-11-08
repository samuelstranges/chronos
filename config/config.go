package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// StorageType represents the type of storage backend to use
type StorageType string

const (
	StorageTypeFile   StorageType = "file"
	StorageTypeCalDAV StorageType = "caldav"
)

// CalDAVConfig holds CalDAV server connection settings
type CalDAVConfig struct {
	ServerURL    string `json:"server_url"`
	Username     string `json:"username"`

	// Password can be specified in multiple ways (priority order):
	// 1. password_command - Execute shell command to get password (e.g., "pass show chronos/caldav")
	// 2. password_env - Read from environment variable (e.g., "CHRONOS_CALDAV_PASSWORD")
	// 3. password - Plain text password (not recommended, but simplest)
	PasswordCommand string `json:"password_command,omitempty"`
	PasswordEnv     string `json:"password_env,omitempty"`
	Password        string `json:"password,omitempty"`

	// CalendarHomeURL - Full URL to calendar home (will discover all calendars under this)
	// e.g., "http://localhost:8086/calendars/user@example.com/"
	CalendarHomeURL string `json:"calendar_home_url"`

	// SyncInterval - Seconds between background syncs (0 = disabled, default: 300 = 5 minutes)
	SyncInterval int `json:"sync_interval,omitempty"`
}

// Config represents the main chronos configuration
type Config struct {
	// Storage backend selection
	StorageType StorageType `json:"storage_type"` // "file" or "caldav"

	// CalDAV configuration (only used if storage_type = "caldav")
	CalDAV *CalDAVConfig `json:"caldav,omitempty"`
}

// Default returns a default configuration (file storage)
func Default() *Config {
	return &Config{
		StorageType: StorageTypeFile,
		CalDAV: &CalDAVConfig{
			ServerURL:       "",
			Username:        "",
			Password:        "",
			CalendarHomeURL: "",
		},
	}
}

// GetConfigPath returns the path to the config file
// Checks both ~/.config/chronos (priority) and system config dir
func GetConfigPath() string {
	// Priority 1: ~/.config/chronos/config.json (XDG standard on Linux/Unix)
	homeDir, err := os.UserHomeDir()
	if err == nil {
		xdgPath := filepath.Join(homeDir, ".config", "chronos", "config.json")
		if _, err := os.Stat(xdgPath); err == nil {
			return xdgPath
		}
	}

	// Priority 2: System config dir (~/Library/Application Support on macOS)
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(configDir, "chronos", "config.json")
}

// Load loads configuration from disk
func Load() (*Config, error) {
	configPath := GetConfigPath()

	// If config doesn't exist, return default
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return Default(), nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate storage type
	if config.StorageType != StorageTypeFile && config.StorageType != StorageTypeCalDAV {
		return nil, fmt.Errorf("invalid storage_type: %s (must be 'file' or 'caldav')", config.StorageType)
	}

	// Validate CalDAV config if needed
	if config.StorageType == StorageTypeCalDAV {
		if config.CalDAV == nil {
			return nil, fmt.Errorf("caldav configuration required when storage_type is 'caldav'")
		}
		if err := validateCalDAVConfig(config.CalDAV); err != nil {
			return nil, fmt.Errorf("invalid caldav config: %w", err)
		}
	}

	return &config, nil
}

// Save saves configuration to disk
func Save(config *Config) error {
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

	// Write to file (0600 for security - contains credentials)
	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GetPassword resolves the password from various sources
// Priority: password_command > password_env > password
func (c *CalDAVConfig) GetPassword() (string, error) {
	// 1. Try password_command (highest priority)
	if c.PasswordCommand != "" {
		password, err := executePasswordCommand(c.PasswordCommand)
		if err != nil {
			return "", fmt.Errorf("failed to execute password_command: %w", err)
		}
		return password, nil
	}

	// 2. Try password_env
	if c.PasswordEnv != "" {
		password := os.Getenv(c.PasswordEnv)
		if password == "" {
			return "", fmt.Errorf("environment variable %s is not set or empty", c.PasswordEnv)
		}
		return password, nil
	}

	// 3. Fall back to plain text password
	if c.Password != "" {
		return c.Password, nil
	}

	return "", fmt.Errorf("no password configured (set password_command, password_env, or password)")
}

// executePasswordCommand executes a shell command and returns its output as password
func executePasswordCommand(command string) (string, error) {
	// Execute command using sh -c to support shell features (pipes, etc.)
	cmd := exec.Command("sh", "-c", command)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("command failed: %w", err)
	}

	// Trim whitespace and newlines from output
	password := strings.TrimSpace(string(output))

	if password == "" {
		return "", fmt.Errorf("command returned empty output")
	}

	return password, nil
}

// validateCalDAVConfig validates CalDAV configuration
func validateCalDAVConfig(c *CalDAVConfig) error {
	if c.ServerURL == "" {
		return fmt.Errorf("server_url is required")
	}

	if c.Username == "" {
		return fmt.Errorf("username is required")
	}

	// Validate that at least one password source is configured
	if c.PasswordCommand == "" && c.PasswordEnv == "" && c.Password == "" {
		return fmt.Errorf("password required (set password_command, password_env, or password)")
	}

	// Try to resolve password to ensure it works
	_, err := c.GetPassword()
	if err != nil {
		return fmt.Errorf("failed to resolve password: %w", err)
	}

	return nil
}
