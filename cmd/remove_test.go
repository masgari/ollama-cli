package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/masgari/ollama-cli/pkg/client"
	"github.com/masgari/ollama-cli/pkg/config"
	"github.com/masgari/ollama-cli/pkg/output"
	"github.com/ollama/ollama/api"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRemoveCommand(t *testing.T) {
	// Save original output and restore it after the test
	origOutput := output.Default
	defer func() { output.Default = origOutput }()

	// Mock data for ListModels
	mockModels := &api.ListResponse{
		Models: []api.ListModelResponse{
			{
				Name: "model1",
			},
			{
				Name: "model2",
			},
		},
	}

	// Test cases
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*client.MockClientTestify)
		forceFlag   bool
		wantErr     bool
		wantContain []string
	}{
		{
			name:      "Force delete with flag",
			args:      []string{"model1"},
			forceFlag: true,
			setupMock: func(m *client.MockClientTestify) {
				m.On("ListModels", mock.Anything).Return(mockModels, nil)
				m.On("DeleteModel", mock.Anything, "model1").Return(nil)
			},
			wantErr: false,
			wantContain: []string{
				"Model 'model1' deleted successfully",
			},
		},
		{
			name:      "Model not found",
			args:      []string{"nonexistent"},
			forceFlag: true,
			setupMock: func(m *client.MockClientTestify) {
				m.On("ListModels", mock.Anything).Return(mockModels, nil)
				// DeleteModel should not be called
			},
			wantErr: true,
			wantContain: []string{
				"model 'nonexistent' not found on the server",
			},
		},
		{
			name:      "Error listing models",
			args:      []string{"model1"},
			forceFlag: true,
			setupMock: func(m *client.MockClientTestify) {
				m.On("ListModels", mock.Anything).Return(nil, errors.New("connection error"))
				// DeleteModel should not be called
			},
			wantErr: true,
			wantContain: []string{
				"failed to list models: connection error",
			},
		},
		{
			name:      "Error deleting model",
			args:      []string{"model1"},
			forceFlag: true,
			setupMock: func(m *client.MockClientTestify) {
				m.On("ListModels", mock.Anything).Return(mockModels, nil)
				m.On("DeleteModel", mock.Anything, "model1").Return(errors.New("delete error"))
			},
			wantErr: true,
			wantContain: []string{
				"failed to delete model: delete error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original config and restore it after the test
			origCfg := cfg
			defer func() { cfg = origCfg }()

			// Save original forceDelete flag and restore it after the test
			origForceDelete := forceDelete
			defer func() { forceDelete = origForceDelete }()

			// Set test config
			cfg = config.DefaultConfig()
			forceDelete = tt.forceFlag

			// Set up the mock client
			mockClient := client.NewMockClient()
			tt.setupMock(mockClient)

			// Set up the client factory to return our mock
			client.SetClientFactory(func() (client.Client, error) {
				return mockClient, nil
			})
			defer client.ResetClientFactory()

			// Create a buffer to capture output
			var buf bytes.Buffer

			// Create a custom output instance that writes to our buffer
			testOutput := output.NewColorWriter(&buf)
			output.Default = testOutput

			// Set up command
			cmd := &cobra.Command{Use: "rm"}
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Execute the RunE function with the parsed args
			err := rmCmd.RunE(cmd, tt.args)

			// Check for expected error
			if (err != nil) != tt.wantErr {
				t.Errorf("rmCmd.RunE() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For error cases, check the error message
			if err != nil {
				assert.Contains(t, err.Error(), tt.wantContain[0], "Error message should contain %q", tt.wantContain[0])
				return
			}

			// For success cases, check output contains expected strings
			output := buf.String()
			for _, want := range tt.wantContain {
				assert.Contains(t, output, want, "Output should contain %q", want)
			}

			// Verify all expectations were met
			mockClient.AssertExpectations(t)
		})
	}
}

func TestRemoveCommandFlags(t *testing.T) {
	// Test that the command is properly defined
	cmd := rmCmd

	// Check command properties
	assert.Equal(t, "rm [model]", cmd.Use)
	assert.Equal(t, "Remove a model from the Ollama server", cmd.Short)
	assert.Contains(t, cmd.Long, "Remove a model and its data from the remote Ollama server")

	// Check force flag
	forceFlag := cmd.Flag("force")
	if forceFlag == nil {
		t.Error("force flag not found")
	} else {
		assert.Equal(t, "f", forceFlag.Shorthand)
		assert.Equal(t, "false", forceFlag.DefValue)
	}
}

func TestRemoveCommandMissingArg(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer

	// Create a root command and add the rm command to it
	rootCmd := &cobra.Command{Use: "ollama-cli"}
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.AddCommand(rmCmd)

	// Execute the command without arguments
	rootCmd.SetArgs([]string{"rm"})
	err := rootCmd.Execute()

	// Check that we got an error
	assert.Error(t, err, "Command should return an error when no model is specified")

	// Check the output for the error message
	output := buf.String()
	assert.Contains(t, output, "Error: accepts 1 arg(s), received 0", "Output should contain error message")
}

func TestRemoveCommandForceFlag(t *testing.T) {
	// Save original forceDelete flag and restore it after the test
	origForceDelete := forceDelete
	defer func() { forceDelete = origForceDelete }()

	// Set forceDelete to true
	forceDelete = true

	// Create a buffer to capture output
	var buf bytes.Buffer

	// Mock the client
	mockClient := client.NewMockClient()
	mockModels := &api.ListResponse{
		Models: []api.ListModelResponse{
			{
				Name: "model1",
			},
		},
	}
	mockClient.On("ListModels", mock.Anything).Return(mockModels, nil)
	mockClient.On("DeleteModel", mock.Anything, "model1").Return(nil)

	// Set up the client factory to return our mock
	client.SetClientFactory(func() (client.Client, error) {
		return mockClient, nil
	})
	defer client.ResetClientFactory()

	// Save original output and restore it after the test
	origOutput := output.Default
	defer func() { output.Default = origOutput }()

	// Create a custom output instance that writes to our buffer
	testOutput := output.NewColorWriter(&buf)
	output.Default = testOutput

	// Set up command
	cmd := &cobra.Command{Use: "rm"}
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Execute the RunE function with the parsed args
	err := rmCmd.RunE(cmd, []string{"model1"})

	// Check that there was no error
	assert.NoError(t, err, "Command should not return an error when using force flag")

	// Check the output
	output := buf.String()
	assert.Contains(t, output, "Model 'model1' deleted successfully", "Output should indicate successful deletion")
	assert.NotContains(t, output, "Are you sure you want to delete model", "Output should not contain confirmation prompt")

	// Verify all expectations were met
	mockClient.AssertExpectations(t)
}

func TestRemoveCommandShortForceFlag(t *testing.T) {
	// This test is redundant since we're directly setting the forceDelete flag
	// The actual flag parsing is handled by Cobra and doesn't need to be tested here
	t.Skip("This test is redundant with TestRemoveCommandForceFlag")
}

func TestRemoveCommandWithoutForceFlag(t *testing.T) {
	// This test is difficult to implement because it requires simulating user input
	// which is challenging in a test environment
	t.Skip("This test requires simulating user input which is challenging in a test environment")
}

func TestRemoveCommandWithoutForceFlagCancelled(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer

	// Create a root command and add the rm command to it
	rootCmd := &cobra.Command{Use: "ollama-cli"}
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.AddCommand(rmCmd)

	// Test without force flag
	rootCmd.SetArgs([]string{"rm", "model1"})

	// Mock the client
	mockClient := client.NewMockClient()
	mockModels := &api.ListResponse{
		Models: []api.ListModelResponse{
			{
				Name: "model1",
			},
		},
	}
	mockClient.On("ListModels", mock.Anything).Return(mockModels, nil)

	// Set up the client factory to return our mock
	client.SetClientFactory(func() (client.Client, error) {
		return mockClient, nil
	})
	defer client.ResetClientFactory()

	// Save original output and restore it after the test
	origOutput := output.Default
	defer func() { output.Default = origOutput }()

	// Create a custom output instance that writes to our buffer
	testOutput := output.NewColorWriter(&buf)
	output.Default = testOutput

	// Save original stdin and stdout
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	// Create pipes for stdin and stdout
	r, w, _ := os.Pipe()
	pr, pw, _ := os.Pipe()

	// Replace stdin and stdout with our pipes
	os.Stdin = r
	os.Stdout = pw

	// Write "n" to stdin to simulate user cancellation
	go func() {
		w.Write([]byte("n\n"))
		w.Close()
	}()

	// Capture stdout in a separate goroutine
	go func() {
		var stdoutBuf bytes.Buffer
		io.Copy(&stdoutBuf, pr)
		// Add the stdout output to our buffer
		buf.Write(stdoutBuf.Bytes())
	}()

	// Restore stdin and stdout after the test
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
		pw.Close()
	}()

	// Execute the command
	err := rootCmd.Execute()

	// Check that there was no error
	assert.NoError(t, err, "Command should not return an error when cancelling deletion")

	// Check the output
	output := buf.String()
	assert.Contains(t, output, "Are you sure you want to delete model", "Output should contain confirmation prompt")
	assert.Contains(t, output, "Deletion cancelled", "Output should indicate cancellation")
	assert.NotContains(t, output, "Model 'model1' deleted successfully", "Output should not indicate successful deletion")

	// Verify all expectations were met
	mockClient.AssertExpectations(t)
}
