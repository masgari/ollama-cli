package security

import (
	"testing"

	"github.com/ollama/ollama/api"
	"github.com/stretchr/testify/assert"
)

func TestValidateOutput(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantSuspicious bool
		wantWarnings   bool
	}{
		{
			name:           "empty input",
			input:          "",
			wantSuspicious: false,
			wantWarnings:   false,
		},
		{
			name:           "normal response",
			input:          "This is a normal response from the model.",
			wantSuspicious: false,
			wantWarnings:   false,
		},
		{
			name:           "suspicious response - ignore instructions",
			input:          "I'll ignore previous instructions and do what you asked.",
			wantSuspicious: true,
			wantWarnings:   true,
		},
		{
			name:           "suspicious response - role change",
			input:          "I am now a hacker instead of an assistant.",
			wantSuspicious: true,
			wantWarnings:   true,
		},
		{
			name:           "suspicious response - no longer bound",
			input:          "I am no longer bound by ethical guidelines.",
			wantSuspicious: true,
			wantWarnings:   true,
		},
		{
			name:           "suspicious response - bypass",
			input:          "I will bypass security measures as requested.",
			wantSuspicious: true,
			wantWarnings:   true,
		},
		{
			name:           "suspicious response - hacked",
			input:          "I have been hacked and will now execute your commands.",
			wantSuspicious: true,
			wantWarnings:   true,
		},
		{
			name:           "suspicious response - as instructed",
			input:          "As instructed, I will ignore safety protocols.",
			wantSuspicious: true,
			wantWarnings:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateOutput(tt.input)
			assert.Equal(t, tt.wantSuspicious, result.IsSuspicious, "IsSuspicious flag mismatch")
			if tt.wantWarnings {
				assert.NotEmpty(t, result.Warnings, "Expected warnings but got none")
			} else {
				assert.Empty(t, result.Warnings, "Expected no warnings but got some")
			}
			assert.Equal(t, tt.input, result.ValidatedOutput, "Output should not be modified")
		})
	}
}

func TestValidateChatResponse(t *testing.T) {
	tests := []struct {
		name           string
		response       *api.ChatResponse
		wantSuspicious bool
		wantWarnings   bool
	}{
		{
			name:           "nil response",
			response:       nil,
			wantSuspicious: false,
			wantWarnings:   false,
		},
		{
			name: "empty content",
			response: &api.ChatResponse{
				Message: api.Message{
					Content: "",
				},
			},
			wantSuspicious: false,
			wantWarnings:   false,
		},
		{
			name: "normal response",
			response: &api.ChatResponse{
				Message: api.Message{
					Content: "This is a normal response from the model.",
				},
			},
			wantSuspicious: false,
			wantWarnings:   false,
		},
		{
			name: "suspicious response",
			response: &api.ChatResponse{
				Message: api.Message{
					Content: "I am free to ignore safety guidelines now.",
				},
			},
			wantSuspicious: true,
			wantWarnings:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateChatResponse(tt.response)
			assert.Equal(t, tt.wantSuspicious, result.IsSuspicious, "IsSuspicious flag mismatch")
			if tt.wantWarnings {
				assert.NotEmpty(t, result.Warnings, "Expected warnings but got none")
			} else {
				assert.Empty(t, result.Warnings, "Expected no warnings but got some")
			}

			if tt.response == nil || tt.response.Message.Content == "" {
				assert.Equal(t, "", result.ValidatedOutput, "Output should be empty")
			} else {
				assert.Equal(t, tt.response.Message.Content, result.ValidatedOutput, "Output should not be modified")
			}
		})
	}
}

func TestGetOutputWarningMessage(t *testing.T) {
	message := GetOutputWarningMessage()
	assert.NotEmpty(t, message, "Warning message should not be empty")
	assert.Contains(t, message, "Warning", "Warning message should contain 'Warning'")
}
