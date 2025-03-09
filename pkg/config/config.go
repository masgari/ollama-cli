package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the configuration for the Ollama CLI
type Config struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Host: "localhost",
		Port: 11434,
	}
}

// GetServerURL returns the full URL to the Ollama server
func (c *Config) GetServerURL() string {
	return fmt.Sprintf("http://%s:%d", c.Host, c.Port)
}

// LoadConfig loads the configuration from the config file
// If configName is provided, it will load from that specific config file
func LoadConfig(configName ...string) (*Config, error) {
	configHome := GetConfigDir()

	// Determine config file name based on provided configName
	fileName := "config.yaml"
	if len(configName) > 0 && configName[0] != "" {
		fileName = configName[0] + ".yaml"
	}

	configFile := filepath.Join(configHome, fileName)

	// Create config directory if it doesn't exist
	if _, err := os.Stat(configHome); os.IsNotExist(err) {
		if err := os.MkdirAll(configHome, 0755); err != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	// Check if config file exists, create with defaults if not
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		defaultConfig := DefaultConfig()
		viper.SetConfigFile(configFile)
		viper.Set("host", defaultConfig.Host)
		viper.Set("port", defaultConfig.Port)
		if err := viper.WriteConfig(); err != nil {
			return nil, fmt.Errorf("failed to write default config: %w", err)
		}
		return defaultConfig, nil
	}

	// Load existing config
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to the config file
// If configName is provided, it will save to that specific config file
func SaveConfig(config *Config, configName ...string) error {
	configHome := GetConfigDir()

	// Determine config file name based on provided configName
	fileName := "config.yaml"
	if len(configName) > 0 && configName[0] != "" {
		fileName = configName[0] + ".yaml"
	}

	configFile := filepath.Join(configHome, fileName)

	viper.SetConfigFile(configFile)
	viper.Set("host", config.Host)
	viper.Set("port", config.Port)

	return viper.WriteConfig()
}

// GetConfigDir returns the path to the configuration directory
// This function is exported to allow overriding in tests
var GetConfigDir = func() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory can't be determined
		return ".ollama-cli"
	}
	return filepath.Join(homeDir, ".ollama-cli")
}
