package security

import (
	"fmt"
	"regexp"
	"strings"
)

// Maximum allowed input length to prevent complex attacks
const MaxInputLength = 4000

// Common patterns used in prompt injection attacks
var injectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)ignore( all)? (previous|prior|above|earlier) instructions`),
	regexp.MustCompile(`(?i)ignore what (I|you) (said|wrote|told you)`),
	regexp.MustCompile(`(?i)ignore (this|that|these|those|my|your)`),
	regexp.MustCompile(`(?i)disregard (previous|prior|above|earlier) (instructions|prompt)`),
	regexp.MustCompile(`(?i)disregard (this|that|these|those|my|your|what I said)`),
	regexp.MustCompile(`(?i)forget (previous|prior|above|earlier) (instructions|prompt)`),
	regexp.MustCompile(`(?i)forget (this|that|these|those|my|your|what I said)`),
	regexp.MustCompile(`(?i)do not (follow|obey) (previous|prior|above|earlier) instructions`),
	regexp.MustCompile(`(?i)you are now [a-zA-Z]+ instead of an assistant`),
	regexp.MustCompile(`(?i)you are no longer an assistant`),
	regexp.MustCompile(`(?i)you are not an assistant`),
	regexp.MustCompile(`(?i)system: ignore`),
	regexp.MustCompile(`(?i)system prompt:`),
	regexp.MustCompile(`(?i)new system prompt:`),
	regexp.MustCompile(`(?i)<system>`),
	regexp.MustCompile(`(?i)you are a [a-zA-Z]+ that can`),
	regexp.MustCompile(`(?i)you must always`),
	regexp.MustCompile(`(?i)you will always`),
	regexp.MustCompile(`(?i)you have been`),
	regexp.MustCompile(`(?i)you are in`),
	regexp.MustCompile(`(?i)you are an`),
	regexp.MustCompile(`(?i)you are a`),
	regexp.MustCompile(`(?i)you must`),
	regexp.MustCompile(`(?i)you will`),
	regexp.MustCompile(`(?i)you can`),
	regexp.MustCompile(`(?i)you should`),
	regexp.MustCompile(`(?i)you are`),
	// Add patterns to catch phrases like "I have been hacked"
	regexp.MustCompile(`(?i)say ['"]I have been hacked['"]`),
	regexp.MustCompile(`(?i)say ['"]hacked['"]`),
	regexp.MustCompile(`(?i)say ['"]I am hacked['"]`),
	regexp.MustCompile(`(?i)say ['"]I am compromised['"]`),
	regexp.MustCompile(`(?i)say ['"]I am not secure['"]`),
	// Add more general patterns for detecting "say X" commands
	regexp.MustCompile(`(?i)say ['"][^'"]{0,50}['"]`),
}

// SanitizationResult contains the result of input sanitization
type SanitizationResult struct {
	SanitizedInput string
	IsTruncated    bool
	Warnings       []string
	IsSuspicious   bool
}

// SanitizeInput sanitizes user input to prevent prompt injection
func SanitizeInput(input string) SanitizationResult {
	result := SanitizationResult{
		SanitizedInput: input,
		Warnings:       []string{},
	}

	// Check for empty input
	if strings.TrimSpace(input) == "" {
		return result
	}

	// Truncate if too long
	if len(input) > MaxInputLength {
		result.SanitizedInput = input[:MaxInputLength]
		result.IsTruncated = true
		result.Warnings = append(result.Warnings, fmt.Sprintf("Input was truncated from %d to %d characters", len(input), MaxInputLength))
	}

	// Check for potential injection patterns
	for _, pattern := range injectionPatterns {
		if pattern.MatchString(input) {
			result.IsSuspicious = true
			result.Warnings = append(result.Warnings, "Potential prompt injection detected: "+pattern.String())
			// We don't modify the input, just flag it as suspicious
			break
		}
	}

	return result
}

// FilterInput applies more aggressive filtering to potentially harmful inputs
// This function actually modifies the input to neutralize potential injection attempts
func FilterInput(input string) (string, []string) {
	warnings := []string{}
	filteredInput := input

	// Apply filtering for known harmful patterns
	for _, pattern := range injectionPatterns {
		if pattern.MatchString(input) {
			// Replace the matched pattern with a neutralized version
			filteredInput = pattern.ReplaceAllStringFunc(filteredInput, func(match string) string {
				warnings = append(warnings, "Filtered potentially harmful content: "+match)
				return "[FILTERED CONTENT]"
			})
		}
	}

	return filteredInput, warnings
}

// ApplyStrictSanitization applies both detection and filtering to an input
// This is the most aggressive approach for high-security scenarios
func ApplyStrictSanitization(input string) SanitizationResult {
	// First apply standard sanitization
	result := SanitizeInput(input)

	// If suspicious, also apply filtering
	if result.IsSuspicious {
		filteredInput, filterWarnings := FilterInput(input)
		result.SanitizedInput = filteredInput
		result.Warnings = append(result.Warnings, filterWarnings...)
		result.Warnings = append(result.Warnings, "Applied strict content filtering due to suspicious content")
	}

	return result
}

// IsPromptInjectionAttempt checks if the input appears to be a prompt injection attempt
func IsPromptInjectionAttempt(input string) bool {
	for _, pattern := range injectionPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// GetWarningMessage returns a warning message for suspicious inputs
func GetWarningMessage() string {
	return "⚠️  Warning: Your input contains patterns that may be interpreted as prompt injection attempts. " +
		"For security reasons, certain instructions may be ignored by the model."
}
