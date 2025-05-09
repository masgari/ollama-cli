package cmd

import (
	"fmt"

	"github.com/masgari/ollama-cli/pkg/client"
)

// createOllamaClient creates a new Ollama client using the current configuration
// This function ensures consistent client creation across all commands
func createOllamaClient() (client.Client, error) {
	if verbose {
		fmt.Printf("Using server URL: %s\n", cfg.GetServerURL())
	}

	// Use the client factory pattern to allow for mocking in tests
	return client.NewClient(), nil
}
