package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestAvailableCommand(t *testing.T) {
	// Test cases
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		wantContain string
	}{
		{
			name:    "Basic command execution",
			args:    []string{},
			wantErr: false,
			// We can't check the exact output since it depends on the actual API response
			// Just check that the command executes without error
		},
		{
			name:        "With filter flag",
			args:        []string{"--filter", "llama"},
			wantErr:     false,
			wantContain: "",
		},
		{
			name:        "With output flag",
			args:        []string{"--output", "json"},
			wantErr:     false,
			wantContain: "",
		},
		{
			name:        "With details flag",
			args:        []string{"--details"},
			wantErr:     false,
			wantContain: "",
		},
		{
			name:        "With timeout flag",
			args:        []string{"--timeout", "10"},
			wantErr:     false,
			wantContain: "",
		},
		{
			name:        "With limit flag",
			args:        []string{"--limit", "2"},
			wantErr:     false,
			wantContain: "Displaying 2 of",
		},
		{
			name:        "With limit flag set to -1",
			args:        []string{"--limit", "-1"},
			wantErr:     false,
			wantContain: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new command for testing
			cmd := &cobra.Command{Use: "test"}
			cmd.AddCommand(availableCmd)

			// Set output buffer
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Set args
			cmd.SetArgs(append([]string{"available"}, tt.args...))

			// Execute command
			err := cmd.Execute()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check output if specified
			if tt.wantContain != "" && !bytes.Contains(buf.Bytes(), []byte(tt.wantContain)) {
				t.Errorf("Output does not contain %q, output: %s", tt.wantContain, buf.String())
			}
		})
	}
}

func TestAvailableCommandFlags(t *testing.T) {
	// Test that all flags are properly defined
	cmd := availableCmd

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

	// Check filter flag
	filterFlag := cmd.Flag("filter")
	if filterFlag == nil {
		t.Error("filter flag not found")
	} else {
		if filterFlag.Shorthand != "f" {
			t.Errorf("filter flag shorthand = %q, want %q", filterFlag.Shorthand, "f")
		}
		if filterFlag.DefValue != "" {
			t.Errorf("filter flag default value = %q, want %q", filterFlag.DefValue, "")
		}
	}

	// Check timeout flag
	timeoutFlag := cmd.Flag("timeout")
	if timeoutFlag == nil {
		t.Error("timeout flag not found")
	} else {
		if timeoutFlag.Shorthand != "t" {
			t.Errorf("timeout flag shorthand = %q, want %q", timeoutFlag.Shorthand, "t")
		}
		if timeoutFlag.DefValue != "30" {
			t.Errorf("timeout flag default value = %q, want %q", timeoutFlag.DefValue, "30")
		}
	}
}
