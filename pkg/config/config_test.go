package config

import (
	"testing"
)

func TestConfigWithHeaders(t *testing.T) {
	// Create a temporary config directory
	tempDir := t.TempDir()
	originalGetConfigDir := GetConfigDir
	GetConfigDir = func() string {
		return tempDir
	}
	defer func() {
		GetConfigDir = originalGetConfigDir
	}()

	// Test loading config with headers
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify default headers is initialized as empty map
	if config.Headers == nil {
		t.Error("Expected Headers to be initialized as empty map")
	}

	// Test saving config with headers
	config.Headers = map[string]string{
		"Authorization":   "Bearer test-token",
		"X-Custom-Header": "custom-value",
	}

	err = SaveConfig(config)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Test loading the saved config
	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	// Verify headers were saved and loaded correctly
	if len(loadedConfig.Headers) != 2 {
		t.Errorf("Expected 2 headers, got %d", len(loadedConfig.Headers))
	}

	if loadedConfig.Headers["Authorization"] != "Bearer test-token" {
		t.Errorf("Expected Authorization header 'Bearer test-token', got '%s'", loadedConfig.Headers["Authorization"])
	}

	if loadedConfig.Headers["X-Custom-Header"] != "custom-value" {
		t.Errorf("Expected X-Custom-Header 'custom-value', got '%s'", loadedConfig.Headers["X-Custom-Header"])
	}
}

func TestDefaultConfigHeaders(t *testing.T) {
	config := DefaultConfig()

	if config.Headers == nil {
		t.Error("Expected Headers to be initialized in default config")
	}

	if len(config.Headers) != 0 {
		t.Errorf("Expected empty headers map, got %d headers", len(config.Headers))
	}
}

func TestGetServerURLWithBaseUrl(t *testing.T) {
	tests := []struct {
		name     string
		baseUrl  string
		host     string
		port     int
		path     string
		tls      bool
		expected string
	}{
		{
			name:     "BaseUrl with scheme is used directly",
			baseUrl:  "https://example.com:8080/custom",
			host:     "ignored-host",
			port:     1234,
			path:     "/ignored-path",
			tls:      true,
			expected: "https://example.com:8080/custom",
		},
		{
			name:     "BaseUrl without scheme gets http:// prefix",
			baseUrl:  "example.com",
			host:     "ignored-host",
			port:     1234,
			path:     "/ignored-path",
			tls:      true,
			expected: "http://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				BaseUrl: tt.baseUrl,
				Host:    tt.host,
				Port:    tt.port,
				Path:    tt.path,
				Tls:     tt.tls,
			}

			got := cfg.GetServerURL()
			if got != tt.expected {
				t.Errorf("GetServerURL() = %q, want %q", got, tt.expected)
			}
		})
	}
}
