package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/masgari/ollama-cli/pkg/config"
)

func TestCustomHeaders(t *testing.T) {
	// Create a test server that captures headers
	var capturedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"models":[]}`))
	}))
	defer server.Close()

	// Create config with custom headers
	cfg := &config.Config{
		Host: "localhost",
		Port: 11434,
		Tls:  false,
		Headers: map[string]string{
			"Authorization":   "Bearer test-token",
			"X-Custom-Header": "custom-value",
		},
	}

	// Create client
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Override the server URL to use our test server
	ollamaClient := client.(*OllamaClient)
	ollamaClient.serverURL, _ = url.Parse(server.URL)

	// Make a request
	_, err = client.ListModels(context.Background())
	if err != nil {
		t.Fatalf("Failed to list models: %v", err)
	}

	// Verify headers were sent
	if capturedHeaders == nil {
		t.Fatal("No headers were captured")
	}

	if auth := capturedHeaders.Get("Authorization"); auth != "Bearer test-token" {
		t.Errorf("Expected Authorization header 'Bearer test-token', got '%s'", auth)
	}

	if custom := capturedHeaders.Get("X-Custom-Header"); custom != "custom-value" {
		t.Errorf("Expected X-Custom-Header 'custom-value', got '%s'", custom)
	}
}

func TestNoCustomHeaders(t *testing.T) {
	// Create a test server that captures headers
	var capturedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"models":[]}`))
	}))
	defer server.Close()

	// Create config without custom headers
	cfg := &config.Config{
		Host:    "localhost",
		Port:    11434,
		Tls:     false,
		Headers: make(map[string]string), // Empty headers
	}

	// Create client
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Override the server URL to use our test server
	ollamaClient := client.(*OllamaClient)
	ollamaClient.serverURL, _ = url.Parse(server.URL)

	// Make a request
	_, err = client.ListModels(context.Background())
	if err != nil {
		t.Fatalf("Failed to list models: %v", err)
	}

	// Verify no custom headers were sent
	if capturedHeaders == nil {
		t.Fatal("No headers were captured")
	}

	if auth := capturedHeaders.Get("Authorization"); auth != "" {
		t.Errorf("Expected no Authorization header, got '%s'", auth)
	}
}
