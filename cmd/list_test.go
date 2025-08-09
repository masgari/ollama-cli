package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/masgari/ollama-cli/pkg/client"
	"github.com/masgari/ollama-cli/pkg/config"
	"github.com/ollama/ollama/api"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListCommand(t *testing.T) {
	// Create a fixed time for consistent test output
	fixedTime := time.Date(2024, 3, 8, 12, 0, 0, 0, time.UTC)

	// Test data
	mockModels := &api.ListResponse{
		Models: []api.ListModelResponse{
			{
				Name:       "model1",
				Size:       1024 * 1024 * 1024, // 1GB
				ModifiedAt: fixedTime,
				Details: api.ModelDetails{
					Family:            "llama",
					ParameterSize:     "7B",
					QuantizationLevel: "Q4_0",
				},
			},
			{
				Name:       "model2",
				Size:       2 * 1024 * 1024 * 1024, // 2GB
				ModifiedAt: fixedTime.Add(-24 * time.Hour),
				Details: api.ModelDetails{
					Family:            "mistral",
					ParameterSize:     "7B",
					QuantizationLevel: "Q4_K_M",
				},
			},
		},
	}

	// Save the original time.Now function and restore it after the test
	origTimeNow := timeNow
	defer func() { timeNow = origTimeNow }()

	// Mock the time.Now function to return our fixed time
	timeNow = func() time.Time {
		return fixedTime.Add(365 * 24 * time.Hour) // 1 year later
	}

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
			args: []string{},
			setupMock: func(m *client.MockClientTestify) {
				m.On("ListModels", mock.Anything).Return(mockModels, nil)
			},
			wantErr: false,
			wantContain: []string{
				"model1",
				"model2",
				"1.0 GB",
				"2.0 GB",
				"12 months ago",
			},
		},
		{
			name: "JSON output",
			args: []string{"--output", "json"},
			setupMock: func(m *client.MockClientTestify) {
				m.On("ListModels", mock.Anything).Return(mockModels, nil)
			},
			wantErr: false,
			wantContain: []string{
				`"name": "model1"`,
				`"name": "model2"`,
				`"family": "llama"`,
				`"family": "mistral"`,
			},
		},
		{
			name: "Wide output",
			args: []string{"--output", "wide"},
			setupMock: func(m *client.MockClientTestify) {
				m.On("ListModels", mock.Anything).Return(mockModels, nil)
			},
			wantErr: false,
			wantContain: []string{
				"model1",
				"model2",
				"1.0 GB",
				"2.0 GB",
				"llama",
				"mistral",
				"7B",
				"Q4_0",
				"Q4_K_M",
			},
		},
		{
			name: "Details flag",
			args: []string{"--details"},
			setupMock: func(m *client.MockClientTestify) {
				m.On("ListModels", mock.Anything).Return(mockModels, nil)
			},
			wantErr: false,
			wantContain: []string{
				"model1",
				"model2",
				"1.0 GB",
				"2.0 GB",
				"llama",
				"mistral",
				"7B",
				"Q4_0",
				"Q4_K_M",
			},
		},
		{
			name: "Error from client",
			args: []string{},
			setupMock: func(m *client.MockClientTestify) {
				m.On("ListModels", mock.Anything).Return(nil, errors.New("connection error"))
			},
			wantErr: true,
			wantContain: []string{
				"failed to list models: connection error",
			},
		},
		{
			name: "Invalid output format",
			args: []string{"--output", "invalid"},
			setupMock: func(m *client.MockClientTestify) {
				m.On("ListModels", mock.Anything).Return(mockModels, nil)
			},
			wantErr: true,
			wantContain: []string{
				"invalid output format: invalid",
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

			// Set up command flags
			cmd := &cobra.Command{Use: "list"}
			cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, wide, json)")
			cmd.Flags().BoolVarP(&showDetails, "details", "d", false, "Show detailed information about models")
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Parse flags
			cmd.SetArgs(tt.args)
			err := cmd.ParseFlags(tt.args)
			if err != nil {
				t.Fatalf("Failed to parse flags: %v", err)
			}

			// Execute the command's RunE function directly
			err = listCmd.RunE(cmd, []string{})

			// Check for expected error
			if (err != nil) != tt.wantErr {
				t.Errorf("listCmd.RunE() error = %v, wantErr %v", err, tt.wantErr)
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

func TestListCommandFlags(t *testing.T) {
	// Test that all flags are properly defined
	cmd := listCmd

	// Check output flag
	outputFlag := cmd.Flag("output")
	if outputFlag == nil {
		t.Error("output flag not found")
	} else {
		if outputFlag.Shorthand != "o" {
			t.Errorf("output flag shorthand = %q, want %q", outputFlag.Shorthand, "o")
		}
		if outputFlag.DefValue != "table" {
			t.Errorf("output flag default value = %q, want %q", outputFlag.DefValue, "table")
		}
	}

	// Check details flag
	detailsFlag := cmd.Flag("details")
	if detailsFlag == nil {
		t.Error("details flag not found")
	} else {
		if detailsFlag.Shorthand != "d" {
			t.Errorf("details flag shorthand = %q, want %q", detailsFlag.Shorthand, "d")
		}
		if detailsFlag.DefValue != "false" {
			t.Errorf("details flag default value = %q, want %q", detailsFlag.DefValue, "false")
		}
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want string
	}{
		{
			name: "Bytes",
			size: 500,
			want: "500 B",
		},
		{
			name: "Kilobytes",
			size: 1024 * 500,
			want: "500.0 KB",
		},
		{
			name: "Megabytes",
			size: 1024 * 1024 * 500,
			want: "500.0 MB",
		},
		{
			name: "Gigabytes",
			size: 1024 * 1024 * 1024 * 2,
			want: "2.0 GB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSize(tt.size)
			// Remove ANSI color codes for comparison
			got = strings.ReplaceAll(got, "\033[36m", "")
			got = strings.ReplaceAll(got, "\033[0m", "")
			if got != tt.want {
				t.Errorf("formatSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	// Create a fixed "now" time for testing
	now := time.Date(2024, 3, 9, 6, 50, 23, 0, time.UTC)

	// Save the original time.Now function and restore it after the test
	origTimeNow := timeNow
	defer func() { timeNow = origTimeNow }()

	// Mock the time.Now function to return our fixed time
	timeNow = func() time.Time {
		return now
	}

	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "Minutes ago",
			time: now.Add(-5 * time.Minute),
			want: "ago",
		},
		{
			name: "Hours ago",
			time: now.Add(-2 * time.Hour),
			want: "ago",
		},
		{
			name: "Days ago",
			time: now.Add(-2 * 24 * time.Hour),
			want: "ago",
		},
		{
			name: "Months ago",
			time: now.Add(-2 * 30 * 24 * time.Hour),
			want: "ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTime(tt.time)
			if !strings.Contains(got, tt.want) {
				t.Errorf("formatTime() = %v, want string containing %q", got, tt.want)
			}
		})
	}
}

func TestGetOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		defaultValue string
		want         string
	}{
		{
			name:         "Empty value",
			value:        "",
			defaultValue: "N/A",
			want:         "N/A",
		},
		{
			name:         "Non-empty value",
			value:        "test",
			defaultValue: "N/A",
			want:         "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getOrDefault(tt.value, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}
