package cmd

import (
	"bytes"
	"context"
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

// mockChatClient is a mock implementation of the Client interface for testing
type mockChatClient struct {
	chatResponse    *api.ChatResponse
	chatError       error
	streamResponses []api.ChatResponse
	streamDelay     time.Duration
}

func (m *mockChatClient) ListModels(ctx context.Context) (*api.ListResponse, error) {
	return nil, nil
}

func (m *mockChatClient) GetModelDetails(ctx context.Context, modelName string) (*api.ShowResponse, error) {
	return nil, nil
}

func (m *mockChatClient) DeleteModel(ctx context.Context, modelName string) error {
	return nil
}

func (m *mockChatClient) PullModel(ctx context.Context, modelName string) error {
	return nil
}

func (m *mockChatClient) ChatWithModel(ctx context.Context, modelName string, messages []api.Message, stream bool, options map[string]interface{}) (*api.ChatResponse, error) {
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
			return &finalResp, m.chatError
		}
	}

	return m.chatResponse, m.chatError
}

func TestChatCommand(t *testing.T) {
	// Save the original stdout and restore it after the test
	oldStdout := os.Stdout
	defer func() {
		os.Stdout = oldStdout
	}()

	// Save the original factory and restore it after the test
	defer client.ResetClientFactory()

	// Create a mock client
	mockClient := &mockChatClient{
		chatResponse: &api.ChatResponse{
			Message: api.Message{
				Role:    "assistant",
				Content: "This is a test response",
			},
		},
	}

	// Set the client factory to return our mock client
	client.SetClientFactory(func() (client.Client, error) {
		return mockClient, nil
	})

	// Create a new command
	cmd := &cobra.Command{Use: "test"}
	cmd.AddCommand(chatCmd)

	// Enable chat for testing
	if config.Current == nil {
		config.Current = &config.Config{
			Host:        "localhost",
			Port:        11434,
			ChatEnabled: true,
		}
	} else {
		config.Current.ChatEnabled = true
	}

	// Test with basic prompt
	t.Run("Basic chat with prompt flag", func(t *testing.T) {
		// Create a pipe to capture stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Set up command line arguments
		cmd.SetArgs([]string{"chat", "test-model", "--prompt", "Hello", "--no-stream"})

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
		assert.Contains(t, output, "This is a test response")
	})

	// Test with system prompt
	t.Run("Chat with system prompt", func(t *testing.T) {
		// Create a pipe to capture stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Set up command line arguments
		cmd.SetArgs([]string{"chat", "test-model", "--prompt", "Hello", "--system", "You are a test assistant", "--no-stream"})

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
		assert.Contains(t, output, "This is a test response")
	})

	// Test with output file
	t.Run("Chat with output file", func(t *testing.T) {
		// Create a temporary file
		tmpfile, err := os.CreateTemp("", "chat-test-*.json")
		assert.NoError(t, err)
		defer os.Remove(tmpfile.Name())
		tmpfile.Close()

		// Create a buffer to capture output
		var buf bytes.Buffer
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Set up command line arguments
		cmd.SetArgs([]string{"chat", "test-model", "--prompt", "Hello", "--output-file", tmpfile.Name(), "--no-stream"})

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
		assert.Contains(t, output, "This is a test response")

		// Verify the file was created and contains the expected content
		fileContent, err := os.ReadFile(tmpfile.Name())
		assert.NoError(t, err)
		assert.Contains(t, string(fileContent), "This is a test response")
	})

	// Test with stats flag
	t.Run("Chat with stats flag", func(t *testing.T) {
		// Save the original stdout and stderr
		oldStdout := os.Stdout
		oldStderr := os.Stderr
		defer func() {
			os.Stdout = oldStdout
			os.Stderr = oldStderr
		}()

		// Create pipes to capture stdout and stderr
		rOut, wOut, _ := os.Pipe()
		rErr, wErr, _ := os.Pipe()
		os.Stdout = wOut
		os.Stderr = wErr

		// Set up a mock client with statistics
		mockClient.chatResponse = &api.ChatResponse{
			Message: api.Message{
				Role:    "assistant",
				Content: "This is a test response",
			},
			Done: true,
			Metrics: api.Metrics{
				TotalDuration:      1000000000, // 1 second
				LoadDuration:       100000000,  // 100 ms
				PromptEvalCount:    10,
				PromptEvalDuration: 200000000, // 200 ms
				EvalCount:          20,
				EvalDuration:       700000000, // 700 ms
			},
		}

		// Set up command line arguments
		cmd.SetArgs([]string{"chat", "test-model", "--prompt", "Hello", "--stats", "--no-stream"})

		// Execute the command
		err := cmd.Execute()
		assert.NoError(t, err)

		// Close the write ends of the pipes to flush the buffers
		wOut.Close()
		wErr.Close()

		// Read the captured outputs
		var bufOut, bufErr bytes.Buffer
		io.Copy(&bufOut, rOut)
		io.Copy(&bufErr, rErr)
		stdoutOutput := bufOut.String()
		stderrOutput := bufErr.String()

		t.Logf("Actual stdout: %q", stdoutOutput)
		t.Logf("Actual stderr: %q", stderrOutput)

		// Check that stdout contains the response content
		assert.Contains(t, stdoutOutput, "This is a test response")

		// Check that stderr contains the statistics
		assert.Contains(t, stderrOutput, "Statistics:")
		assert.Contains(t, stderrOutput, "Total time:")
		assert.Contains(t, stderrOutput, "Load time:")
		assert.Contains(t, stderrOutput, "Prompt tokens:")
		assert.Contains(t, stderrOutput, "Response tokens:")
		assert.Contains(t, stderrOutput, "Generation speed:")
	})
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		durationMs float64
		expected   string
	}{
		{0.123, "0.123 ms"},
		{0.5, "0.500 ms"},
		{1.0, "1.00 ms"},
		{10.5, "10.50 ms"},
		{999.99, "999.99 ms"},
		{1000.0, "1.00 s"},
		{1500.0, "1.50 s"},
		{9999.0, "10.00 s"},
		{30000.0, "30.0 s"},
		{59999.0, "60.0 s"},
		{60000.0, "1 min"},
		{61000.0, "1 min 1.0 s"},
		{90000.0, "1 min 30.0 s"},
		{120000.0, "2 min"},
		{3599999.0, "60 min"},
		{3600000.0, "1 h"},
		{3660000.0, "1 h 1 min"},
		{7200000.0, "2 h"},
		{86399999.0, "23 h 59 min"},
		{86400000.0, "1 days"},
		{90000000.0, "1 days 1 h"},
		{172800000.0, "2 days"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%.2f ms", test.durationMs), func(t *testing.T) {
			result := formatDuration(test.durationMs)
			assert.Equal(t, test.expected, result, "Formatting %f ms", test.durationMs)
		})
	}
}
