package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config struct for storing database connection information
type Config struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"sslmode"`
	// K8s related configuration
	Namespace string `json:"namespace,omitempty"`
	Pod       string `json:"pod,omitempty"`
	Container string `json:"container,omitempty"`
	PortName  string `json:"port_name,omitempty"` // Save selected port name
	Secret    string `json:"secret,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
}

// getConfigPath returns the path of config file
func getConfigPath() (string, error) {
	// Get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to get user home directory: %v", err)
	}

	// Create .p6s directory if it doesn't exist
	configDir := filepath.Join(homeDir, ".p6s")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("unable to create config directory: %v", err)
	}

	// Return full path of config file
	return filepath.Join(configDir, "config.json"), nil
}

// LoadConfig loads config from config file
func LoadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", configPath)
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse JSON data
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

// SaveConfig saves config to config file
func SaveConfig(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Convert config to JSON data
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %v", err)
	}

	// Write to config file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// BuildConnStr builds connection string (read-only mode)
func BuildConnStr(host, port, username, password, database, sslmode string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&default_transaction_read_only=on&application_name=p6s-readonly",
		username, password, host, port, database, sslmode)
}