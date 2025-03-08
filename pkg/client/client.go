package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/masgari/ollama-cli/pkg/config"
	"github.com/ollama/ollama/api"
)

// Client represents an Ollama API client
type Client struct {
	apiClient *api.Client
	config    *config.Config
}

// New creates a new Ollama client
func New(cfg *config.Config) (*Client, error) {
	serverURL, err := url.Parse(cfg.GetServerURL())
	if err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}

	// Create a new HTTP client with a timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	apiClient := api.NewClient(serverURL, httpClient)

	return &Client{
		apiClient: apiClient,
		config:    cfg,
	}, nil
}

// ListModels lists all models available on the Ollama server
func (c *Client) ListModels(ctx context.Context) (*api.ListResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	models, err := c.apiClient.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	return models, nil
}

// GetModelDetails gets details for a specific model
func (c *Client) GetModelDetails(ctx context.Context, modelName string) (*api.ShowResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req := &api.ShowRequest{
		Model: modelName,
	}

	model, err := c.apiClient.Show(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get model details: %w", err)
	}

	return model, nil
}

// DeleteModel deletes a model from the Ollama server
func (c *Client) DeleteModel(ctx context.Context, modelName string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req := &api.DeleteRequest{
		Model: modelName,
	}

	if err := c.apiClient.Delete(ctx, req); err != nil {
		return fmt.Errorf("failed to delete model: %w", err)
	}

	return nil
}
