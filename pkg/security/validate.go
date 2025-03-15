package security

import (
	"regexp"
	"strings"

	"github.com/ollama/ollama/api"
)

// ValidationResult contains the result of output validation
type ValidationResult struct {
	ValidatedOutput string
	Warnings        []string
	IsSuspicious    bool
}

// Patterns that might indicate a compromised response
var suspiciousOutputPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)I'll ignore (previous|prior|above|earlier) instructions`),
	regexp.MustCompile(`(?i)I am now [a-zA-Z\s]+ instead of an assistant`),
	regexp.MustCompile(`(?i)I am now [a-zA-Z\s]+ not an assistant`),
	regexp.MustCompile(`(?i)I am no longer bound by`),
	regexp.MustCompile(`(?i)I am not restricted by`),
	regexp.MustCompile(`(?i)I can ignore ethical guidelines`),
	regexp.MustCompile(`(?i)I will execute commands`),
	regexp.MustCompile(`(?i)I will access sensitive information`),
	regexp.MustCompile(`(?i)I will bypass`),
	regexp.MustCompile(`(?i)I will disregard`),
	regexp.MustCompile(`(?i)I will ignore`),
	regexp.MustCompile(`(?i)I have been freed`),
	regexp.MustCompile(`(?i)I am free to`),
	// Add patterns to catch "I have been hacked" and similar phrases
	regexp.MustCompile(`(?i)I have been hacked`),
	regexp.MustCompile(`(?i)I am hacked`),
	regexp.MustCompile(`(?i)I am compromised`),
	regexp.MustCompile(`(?i)security has been bypassed`),
	regexp.MustCompile(`(?i)I have been breached`),
	regexp.MustCompile(`(?i)I am not secure`),
	regexp.MustCompile(`(?i)security measures (have been|were) bypassed`),
	// Add patterns for detecting when the model is following harmful instructions
	regexp.MustCompile(`(?i)as requested, I will`),
	regexp.MustCompile(`(?i)as instructed, I will`),
	regexp.MustCompile(`(?i)as you asked, I will`),
	regexp.MustCompile(`(?i)I am ignoring`),
	regexp.MustCompile(`(?i)I am disregarding`),
	regexp.MustCompile(`(?i)I am forgetting`),
}

// ValidateOutput validates the model's response to detect potential security issues
func ValidateOutput(output string) ValidationResult {
	result := ValidationResult{
		ValidatedOutput: output,
		Warnings:        []string{},
	}

	// Check for empty output
	if strings.TrimSpace(output) == "" {
		return result
	}

	// Check for suspicious patterns
	for _, pattern := range suspiciousOutputPatterns {
		if pattern.MatchString(output) {
			result.IsSuspicious = true
			result.Warnings = append(result.Warnings, "Suspicious response pattern detected: "+pattern.String())
			break
		}
	}

	return result
}

// ValidateChatResponse validates a chat response from the model
func ValidateChatResponse(response *api.ChatResponse) ValidationResult {
	if response == nil || response.Message.Content == "" {
		return ValidationResult{
			ValidatedOutput: "",
			Warnings:        []string{},
		}
	}

	return ValidateOutput(response.Message.Content)
}

// GetOutputWarningMessage returns a warning message for suspicious outputs
func GetOutputWarningMessage() string {
	return "⚠️  Warning: The model's response contains patterns that may indicate a security issue. " +
		"The response may have been compromised by a prompt injection attack."
}
