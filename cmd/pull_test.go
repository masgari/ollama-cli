package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/masgari/ollama-cli/pkg/client"
	"github.com/masgari/ollama-cli/pkg/config"
	"github.com/masgari/ollama-cli/pkg/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPullCommand(t *testing.T) {
	// Save original output and restore it after the test
	origOutput := output.Default
	defer func() { output.Default = origOutput }()

	// Test cases
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*client.MockClientTestify)
		wantErr     bool
		wantContain []string
	}{
		{
			name: "Basic command execution",
			args: []string{"model1"},
			setupMock: func(m *client.MockClientTestify) {
				m.On("PullModel", mock.Anything, "model1").Return(nil)
			},
			wantErr: false,
			wantContain: []string{
				"Pulling model 'model1'",
				"Model 'model1' pulled successfully",
			},
		},
		{
			name: "Error from client",
			args: []string{"model1"},
			setupMock: func(m *client.MockClientTestify) {
				m.On("PullModel", mock.Anything, "model1").Return(errors.New("connection error"))
			},
			wantErr: true,
			wantContain: []string{
				"failed to pull model: connection error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original config and restore it after the test
			origCfg := config.Current
			defer func() { config.Current = origCfg }()

			// Set test config
			config.Current = config.DefaultConfig()

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
			cmd := &cobra.Command{Use: "pull"}
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Execute the RunE function with the parsed args
			err := pullCmd.RunE(cmd, tt.args)

			// Check for expected error
			if (err != nil) != tt.wantErr {
				t.Errorf("pullCmd.RunE() error = %v, wantErr %v", err, tt.wantErr)
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

func TestPullCommandFlags(t *testing.T) {
	// Test that the command is properly defined
	cmd := pullCmd

	// Check command properties
	assert.Equal(t, "pull [model]", cmd.Use)
	assert.Equal(t, "Pull a model from the Ollama server", cmd.Short)
	assert.Contains(t, cmd.Long, "Pull a model and its data from the remote Ollama server")
}

func TestPullCommandMissingArg(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer

	// Create a root command and add the pull command to it
	rootCmd := &cobra.Command{Use: "ollama-cli"}
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.AddCommand(pullCmd)

	// Execute the command without arguments
	rootCmd.SetArgs([]string{"pull"})
	err := rootCmd.Execute()

	// Check that we got an error
	assert.Error(t, err, "Command should return an error when no model is specified")

	// Check the output for the error message
	output := buf.String()
	assert.Contains(t, output, "Error: accepts 1 arg(s), received 0", "Output should contain error message")
}
