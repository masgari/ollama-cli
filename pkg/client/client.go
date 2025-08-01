package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"errors"
	"net"

	"github.com/masgari/ollama-cli/pkg/config"
	"github.com/masgari/ollama-cli/pkg/output"
	"github.com/masgari/ollama-cli/pkg/security"
	"github.com/ollama/ollama/api"
)

// Client represents an Ollama API client interface
type Client interface {
	ListModels(ctx context.Context) (*api.ListResponse, error)
	GetModelDetails(ctx context.Context, modelName string) (*api.ShowResponse, error)
	DeleteModel(ctx context.Context, modelName string) error
	PullModel(ctx context.Context, modelName string) error
	ChatWithModel(ctx context.Context, modelName string, messages []api.Message, stream bool, options map[string]interface{}) (*api.ChatResponse, error)
}

// OllamaClient represents an Ollama API client implementation
type OllamaClient struct {
	serverURL *url.URL
	config    *config.Config
}

// clientFactory is a function type that creates a new client
type clientFactory func() (Client, error)

// defaultClientFactory is the default implementation of clientFactory
var defaultClientFactory clientFactory = func() (Client, error) {
	// Load config using the global configuration name
	cfg, err := config.LoadConfig(config.CurrentConfigName)
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

// NewClientWithConfig creates a new client with the provided configuration
func NewClientWithConfig(cfg *config.Config) (Client, error) {
	return New(cfg)
}

// New creates a new Ollama client
func New(cfg *config.Config) (Client, error) {
	serverURL, err := url.Parse(cfg.GetServerURL())
	if err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}

	return &OllamaClient{
		serverURL: serverURL,
		config:    cfg,
	}, nil
}

// createClient creates a new HTTP client with the specified timeout
func (c *OllamaClient) createClient(timeout time.Duration, forPull bool) *api.Client {
	transport := &http.Transport{
		DisableKeepAlives: !forPull, // Enable keep-alive for pull operations
		// Add other necessary transport settings
		MaxIdleConns:       100,
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: false,
	}

	httpClient := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	// Add custom headers to all requests if configured
	if len(c.config.Headers) > 0 {
		httpClient.Transport = &headerTransport{
			base:    transport,
			headers: c.config.Headers,
		}
	}

	return api.NewClient(c.serverURL, httpClient)
}

// headerTransport wraps the base transport to add custom headers
type headerTransport struct {
	base    http.RoundTripper
	headers map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add all configured headers to the request
	for key, value := range t.headers {
		req.Header.Set(key, value)
	}
	return t.base.RoundTrip(req)
}

// ListModels lists all models available on the Ollama server
func (c *OllamaClient) ListModels(ctx context.Context) (*api.ListResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	client := c.createClient(30*time.Second, false)
	models, err := client.List(ctx)
	if err != nil {
		if isTimeoutError(err) {
			return nil, fmt.Errorf("timeout while listing models: %w", err)
		}
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	return models, nil
}

// GetModelDetails gets details for a specific model
func (c *OllamaClient) GetModelDetails(ctx context.Context, modelName string) (*api.ShowResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client := c.createClient(10*time.Second, false)
	req := &api.ShowRequest{
		Model: modelName,
	}

	model, err := client.Show(ctx, req)
	if err != nil {
		if isTimeoutError(err) {
			return nil, fmt.Errorf("timeout while getting model details: %w", err)
		}
		return nil, fmt.Errorf("failed to get model details: %w", err)
	}

	return model, nil
}

// DeleteModel deletes a model from the Ollama server
func (c *OllamaClient) DeleteModel(ctx context.Context, modelName string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client := c.createClient(10*time.Second, false)
	req := &api.DeleteRequest{
		Model: modelName,
	}

	if err := client.Delete(ctx, req); err != nil {
		if isTimeoutError(err) {
			return fmt.Errorf("timeout while deleting model: %w", err)
		}
		return fmt.Errorf("failed to delete model: %w", err)
	}

	return nil
}

// PullModel pulls a model from the Ollama server
func (c *OllamaClient) PullModel(ctx context.Context, modelName string) error {
	// Use a very long timeout for pull operations (4 hours)
	ctx, cancel := context.WithTimeout(ctx, 4*time.Hour)
	defer cancel()

	client := c.createClient(4*time.Hour, true) // Enable keep-alive for pull
	req := &api.PullRequest{
		Name: modelName,
	}

	if err := client.Pull(ctx, req, func(progress api.ProgressResponse) error {
		if progress.Status != "" {
			// Calculate percentage if total is available
			var percentStr string
			var sizeStr string
			if progress.Total > 0 {
				percent := float64(progress.Completed) / float64(progress.Total) * 100
				percentStr = fmt.Sprintf("[%s] ", output.Info(fmt.Sprintf("%.1f%%", percent)))
				sizeStr = fmt.Sprintf("[%s] ", output.Warning(fmt.Sprintf("%.1f/%.1f MB", float64(progress.Completed)/1024/1024, float64(progress.Total)/1024/1024)))
			}

			fmt.Printf("\r%s: %s%s%s", output.Highlight(modelName), percentStr, sizeStr, output.Info(progress.Status))
			if progress.Total > 0 && progress.Completed == progress.Total {
				fmt.Println() // Add newline when complete
			}
		}
		return nil
	}); err != nil {
		if isTimeoutError(err) {
			return fmt.Errorf("timeout while pulling model (operation took longer than 4 hours): %w", err)
		}
		return fmt.Errorf("failed to pull model: %w", err)
	}

	return nil
}

// ChatWithModel sends a chat request to the Ollama server
func (c *OllamaClient) ChatWithModel(ctx context.Context, modelName string, messages []api.Message, stream bool, options map[string]interface{}) (*api.ChatResponse, error) {
	// Use a reasonable timeout for chat operations (2 minutes)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	client := c.createClient(30*time.Minute, false)
	req := &api.ChatRequest{
		Model:    modelName,
		Messages: messages,
		Stream:   &stream,
		Options:  options,
	}

	var finalResponse *api.ChatResponse
	var accumulatedContent string

	err := client.Chat(ctx, req, func(response api.ChatResponse) error {
		if stream {
			// Accumulate the content
			accumulatedContent += response.Message.Content

			// Print the response content as it comes in
			fmt.Print(response.Message.Content)
		}

		if response.Done {
			finalResponse = &response

			// If streaming was enabled, update the final response with the accumulated content
			if stream && finalResponse != nil {
				finalResponse.Message.Content = accumulatedContent
			}
		}

		return nil
	})

	if err != nil {
		if isTimeoutError(err) {
			return nil, fmt.Errorf("timeout while chatting with model: %w", err)
		}
		return nil, fmt.Errorf("failed to chat with model: %w", err)
	}

	if stream {
		fmt.Println() // Add a newline at the end of streaming output

		// If we didn't get a final response with Done=true, create one with the accumulated content
		if finalResponse == nil {
			finalResponse = &api.ChatResponse{
				Message: api.Message{
					Role:    "assistant",
					Content: accumulatedContent,
				},
				Done: true,
			}
		}
	}

	// Validate the response for security issues
	if finalResponse != nil {
		validationResult := security.ValidateChatResponse(finalResponse)

		// Display warnings if any
		for _, warning := range validationResult.Warnings {
			output.Default.WarningPrintf("%s\n", warning)
		}

		// If suspicious, display a warning
		if validationResult.IsSuspicious {
			output.Default.WarningPrintf("%s\n", security.GetOutputWarningMessage())
		}
	}

	return finalResponse, nil
}

// isTimeoutError checks if the error is a timeout error
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	return errors.Is(err, context.DeadlineExceeded)
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

func (c *errorClient) ChatWithModel(ctx context.Context, modelName string, messages []api.Message, stream bool, options map[string]interface{}) (*api.ChatResponse, error) {
	return nil, c.err
}
