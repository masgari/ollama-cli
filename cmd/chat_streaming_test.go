package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/masgari/ollama-cli/pkg/client"
	"github.com/masgari/ollama-cli/pkg/config"
	"github.com/ollama/ollama/api"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// mockStreamingClient is a mock implementation of the Client interface for streaming tests
type mockStreamingClient struct {
	streamResponses []api.ChatResponse
	streamDelay     time.Duration
}

func (m *mockStreamingClient) ListModels(ctx context.Context) (*api.ListResponse, error) {
	return nil, nil
}

func (m *mockStreamingClient) GetModelDetails(ctx context.Context, modelName string) (*api.ShowResponse, error) {
	return nil, nil
}

func (m *mockStreamingClient) DeleteModel(ctx context.Context, modelName string) error {
	return nil
}

func (m *mockStreamingClient) PullModel(ctx context.Context, modelName string) error {
	return nil
}

func (m *mockStreamingClient) ChatWithModel(ctx context.Context, modelName string, messages []api.Message, stream bool, options map[string]interface{}) (*api.ChatResponse, error) {
	if stream && len(m.streamResponses) > 0 {
		// If streaming is enabled and we have stream responses, simulate streaming
		var accumulatedContent string
		for i, resp := range m.streamResponses {
			// Call the callback with each response
			fmt.Print(resp.Message.Content)

			// Accumulate the content
			accumulatedContent += resp.Message.Content

			// Add a delay between responses to simulate real-world streaming
			if i < len(m.streamResponses)-1 && m.streamDelay > 0 {
				time.Sleep(m.streamDelay)
			}
		}

		// Return the final response with accumulated content
		if len(m.streamResponses) > 0 {
			finalResp := m.streamResponses[len(m.streamResponses)-1]
			finalResp.Message.Content = accumulatedContent
			finalResp.Done = true
			return &finalResp, nil
		}
	}

	return nil, fmt.Errorf("no streaming responses available")
}

// TestStreamingChatIsolated tests the streaming functionality in isolation
func TestStreamingChatIsolated(t *testing.T) {
	// Reset the client factory at the beginning of the test
	client.ResetClientFactory()

	// Save the original stdout and restore it after the test
	oldStdout := os.Stdout
	defer func() {
		os.Stdout = oldStdout
	}()

	// Save the original factory and restore it after the test
	defer client.ResetClientFactory()

	// Create a mock client with streaming responses
	mockClient := &mockStreamingClient{
		streamResponses: []api.ChatResponse{
			{
				Message: api.Message{
					Role:    "assistant",
					Content: "This ",
				},
			},
			{
				Message: api.Message{
					Role:    "assistant",
					Content: "is ",
				},
			},
			{
				Message: api.Message{
					Role:    "assistant",
					Content: "a ",
				},
			},
			{
				Message: api.Message{
					Role:    "assistant",
					Content: "streaming ",
				},
			},
			{
				Message: api.Message{
					Role:    "assistant",
					Content: "test ",
				},
			},
			{
				Message: api.Message{
					Role:    "assistant",
					Content: "response",
				},
				Done: true,
				Metrics: api.Metrics{
					TotalDuration: 1000000000, // 1 second
				},
			},
		},
		streamDelay: 50 * time.Millisecond, // 50ms delay between chunks
	}

	// Set the client factory to return our mock client
	client.SetClientFactory(func() (client.Client, error) {
		return mockClient, nil
	})

	// Create a new command
	cmd := &cobra.Command{Use: "test"}
	cmd.AddCommand(chatCmd)

	// Enable chat for testing
	if cfg == nil {
		cfg = &config.Config{
			Host:        "localhost",
			Port:        11434,
			ChatEnabled: true,
		}
	} else {
		cfg.ChatEnabled = true
	}

	// Test with streaming enabled (default)
	t.Run("Basic streaming chat", func(t *testing.T) {
		// Create a pipe to capture stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Set up command line arguments
		cmd.SetArgs([]string{"chat", "test-model", "--prompt", "Hello"})

		// Execute the command
		err := cmd.Execute()
		assert.NoError(t, err)

		// Close the write end of the pipe to flush the buffer
		w.Close()

		// Read the captured output
		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		// Check the output
		assert.Contains(t, output, "This is a streaming test response")
	})

	// Test with streaming to output file
	t.Run("Streaming chat with output file", func(t *testing.T) {
		// Create a temporary file
		tmpfile, err := os.CreateTemp("", "chat-stream-test-*.json")
		assert.NoError(t, err)
		defer os.Remove(tmpfile.Name())
		tmpfile.Close()

		// Create a buffer to capture output
		var buf bytes.Buffer
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Set up command line arguments
		cmd.SetArgs([]string{"chat", "test-model", "--prompt", "Hello", "--output-file", tmpfile.Name()})

		// Execute the command
		err = cmd.Execute()
		assert.NoError(t, err)

		// Close the write end of the pipe to flush the buffer
		w.Close()
		io.Copy(&buf, r)
		os.Stdout = oldStdout
		output := buf.String()

		// Print the actual output for debugging
		t.Logf("Actual output: %q", output)

		// Check that the response is in the output
		assert.Contains(t, output, "This is a streaming test response")

		// Verify the file was created and contains the expected content
		fileContent, err := os.ReadFile(tmpfile.Name())
		assert.NoError(t, err)

		// The file should contain the complete response
		assert.Contains(t, string(fileContent), "This is a streaming test response")

		// Parse the JSON to verify it contains the correct message content
		var messages []api.Message
		err = json.Unmarshal(fileContent, &messages)
		assert.NoError(t, err)

		// Find the assistant message
		var assistantMessage string
		for _, msg := range messages {
			if msg.Role == "assistant" {
				assistantMessage = msg.Content
				break
			}
		}

		assert.Equal(t, "This is a streaming test response", assistantMessage)
	})
}
