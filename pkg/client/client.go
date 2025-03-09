package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/masgari/ollama-cli/pkg/config"
	"github.com/masgari/ollama-cli/pkg/output"
	"github.com/ollama/ollama/api"
)

// Client represents an Ollama API client interface
type Client interface {
	ListModels(ctx context.Context) (*api.ListResponse, error)
	GetModelDetails(ctx context.Context, modelName string) (*api.ShowResponse, error)
	DeleteModel(ctx context.Context, modelName string) error
	PullModel(ctx context.Context, modelName string) error
}

// OllamaClient represents an Ollama API client implementation
type OllamaClient struct {
	apiClient *api.Client
	config    *config.Config
}

// clientFactory is a function type that creates a new client
type clientFactory func() (Client, error)

// defaultClientFactory is the default implementation of clientFactory
var defaultClientFactory clientFactory = func() (Client, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return New(cfg)
}

// currentClientFactory is the current factory function to use
var currentClientFactory = defaultClientFactory

// SetClientFactory sets a custom client factory for testing
func SetClientFactory(factory clientFactory) {
	currentClientFactory = factory
}

// ResetClientFactory resets the client factory to the default
func ResetClientFactory() {
	currentClientFactory = defaultClientFactory
}

// NewClient creates a new client using the current factory
func NewClient() Client {
	client, err := currentClientFactory()
	if err != nil {
		// If there's an error, return a client that will return errors for all operations
		return &errorClient{err: err}
	}
	return client
}

// New creates a new Ollama client
func New(cfg *config.Config) (Client, error) {
	serverURL, err := url.Parse(cfg.GetServerURL())
	if err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}

	// Create a new HTTP client with a timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	apiClient := api.NewClient(serverURL, httpClient)

	return &OllamaClient{
		apiClient: apiClient,
		config:    cfg,
	}, nil
}

// ListModels lists all models available on the Ollama server
func (c *OllamaClient) ListModels(ctx context.Context) (*api.ListResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	models, err := c.apiClient.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	return models, nil
}

// GetModelDetails gets details for a specific model
func (c *OllamaClient) GetModelDetails(ctx context.Context, modelName string) (*api.ShowResponse, error) {
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
func (c *OllamaClient) DeleteModel(ctx context.Context, modelName string) error {
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

// PullModel pulls a model from the Ollama server
func (c *OllamaClient) PullModel(ctx context.Context, modelName string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute) // Longer timeout for model downloads
	defer cancel()

	req := &api.PullRequest{
		Name: modelName,
	}

	if err := c.apiClient.Pull(ctx, req, func(progress api.ProgressResponse) error {
		if progress.Status != "" {
			fmt.Printf("\r%s: %s", output.Highlight(modelName), output.Info(progress.Status))
			if progress.Total > 0 && progress.Completed == progress.Total {
				fmt.Println() // Add newline when complete
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to pull model: %w", err)
	}

	return nil
}

// errorClient is a client that returns errors for all operations
type errorClient struct {
	err error
}

func (c *errorClient) ListModels(ctx context.Context) (*api.ListResponse, error) {
	return nil, c.err
}

func (c *errorClient) GetModelDetails(ctx context.Context, modelName string) (*api.ShowResponse, error) {
	return nil, c.err
}

func (c *errorClient) DeleteModel(ctx context.Context, modelName string) error {
	return c.err
}

func (c *errorClient) PullModel(ctx context.Context, modelName string) error {
	return c.err
}
