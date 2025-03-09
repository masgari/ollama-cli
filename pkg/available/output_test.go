package available

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

func TestOutputJSON(t *testing.T) {
	// Test data
	models := []Model{
		{
			Name:        "llama2",
			Description: "Llama 2 model",
			Size:        "7.0B",
			Pulls:       "1M",
			Tags:        "10",
			Updated:     "1 day ago",
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the function
	err := OutputJSON(models)
	if err != nil {
		t.Fatalf("OutputJSON() error = %v", err)
	}

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output is valid JSON
	var result []map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify the content
	if len(result) != 1 {
		t.Errorf("Expected 1 model in JSON output, got %d", len(result))
	}
	if result[0]["name"] != "llama2" {
		t.Errorf("Expected model name 'llama2', got %v", result[0]["name"])
	}
}

func TestOutputTable(t *testing.T) {
	// Test data
	models := []Model{
		{
			Name:        "llama2",
			Description: "Llama 2 model",
			Size:        "7.0B",
			Pulls:       "1M",
			Tags:        "10",
			Updated:     "1 day ago",
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the function
	err := OutputTable(models, false)
	if err != nil {
		t.Fatalf("OutputTable() error = %v", err)
	}

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output contains expected data
	if !strings.Contains(output, "llama2") {
		t.Errorf("Expected output to contain 'llama2', got: %s", output)
	}
	if !strings.Contains(output, "7.0B") {
		t.Errorf("Expected output to contain '7.0B', got: %s", output)
	}
}

func TestOutputWide(t *testing.T) {
	// Test data
	models := []Model{
		{
			Name:        "llama2",
			Description: "Llama 2 model",
			Size:        "7.0B",
			Pulls:       "1M",
			Tags:        "10",
			Updated:     "1 day ago",
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the function
	err := OutputWide(models)
	if err != nil {
		t.Fatalf("OutputWide() error = %v", err)
	}

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output contains expected data
	if !strings.Contains(output, "llama2") {
		t.Errorf("Expected output to contain 'llama2', got: %s", output)
	}
	if !strings.Contains(output, "Llama 2 model") {
		t.Errorf("Expected output to contain 'Llama 2 model', got: %s", output)
	}
}
