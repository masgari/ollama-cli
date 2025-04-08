package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// CurrentConfigName holds the name of the current configuration
var CurrentConfigName string

// Config holds the configuration for the Ollama CLI
type Config struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	ChatEnabled  bool   `mapstructure:"chat_enabled"`
	CheckUpdates bool   `mapstructure:"check_updates"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Host:         "localhost",
		Port:         11434,
		ChatEnabled:  false, // Chat is disabled by default
		CheckUpdates: true,  // Check for updates by default
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
		viper.Set("chat_enabled", defaultConfig.ChatEnabled)
		viper.Set("check_updates", defaultConfig.CheckUpdates)
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
	viper.Set("chat_enabled", config.ChatEnabled)
	viper.Set("check_updates", config.CheckUpdates)

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

// EnableChat enables the chat feature in the configuration and saves it
func EnableChat(configName ...string) error {
	config, err := LoadConfig(configName...)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	config.ChatEnabled = true

	return SaveConfig(config, configName...)
}

// IsChatEnabled checks if chat is enabled in the configuration
func IsChatEnabled(configName ...string) (bool, error) {
	config, err := LoadConfig(configName...)
	if err != nil {
		return false, fmt.Errorf("failed to load config: %w", err)
	}

	return config.ChatEnabled, nil
}
