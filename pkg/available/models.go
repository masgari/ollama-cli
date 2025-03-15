package available

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Model represents a model available on ollama.com
type Model struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Size        string `json:"size,omitempty"`
	Pulls       string `json:"pulls,omitempty"`
	Tags        string `json:"tags,omitempty"`
	Updated     string `json:"updated,omitempty"`
}

// ModelFetcher is responsible for fetching models from a remote server
// It allows dependency injection for testability
type ModelFetcher struct {
	client *http.Client
	url    string
}

// NewModelFetcher creates a new ModelFetcher with the given HTTP client and URL
func NewModelFetcher(client *http.Client, url string) *ModelFetcher {
	return &ModelFetcher{
		client: client,
		url:    url,
	}
}

// FetchModels fetches the list of available models from the specified URL
func (mf *ModelFetcher) FetchModels(ctx context.Context) ([]Model, error) {
	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", mf.url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", "ollama-cli")

	// Send request
	resp, err := mf.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response
	models, err := parseModels(string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return models, nil
}

// Update the existing FetchModels function to use ModelFetcher
func FetchModels(ctx context.Context, timeout int) ([]Model, error) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	fetcher := NewModelFetcher(client, "https://ollama.com/search")
	return fetcher.FetchModels(ctx)
}

// parseModels parses the HTML response from ollama.com/search
func parseModels(html string) ([]Model, error) {
	var models []Model

	// Regular expression to find model blocks - using a non-greedy pattern and making it work with newlines
	modelBlockRegex := regexp.MustCompile(`(?s)<li x-test-model[^>]*>.*?</li>`)
	modelBlocks := modelBlockRegex.FindAllString(html, -1)

	if len(modelBlocks) == 0 {
		return nil, fmt.Errorf("no models found in response")
	}

	// Regular expressions to extract model information within each block
	titleRegex := regexp.MustCompile(`<span x-test-search-response-title>(.*?)</span>`)
	descRegex := regexp.MustCompile(`<p class="max-w-lg break-words[^>]*>(.*?)</p>`)
	sizeRegex := regexp.MustCompile(`<span[^>]*x-test-size[^>]*>(\d+(?:\.\d+)?[bB])</span>`)
	pullsRegex := regexp.MustCompile(`<span x-test-pull-count[^>]*>([^<]+)</span>`)
	tagsRegex := regexp.MustCompile(`<span x-test-tag-count[^>]*>([^<]+)</span>`)
	updatedRegex := regexp.MustCompile(`<span x-test-updated[^>]*>([^<]+)</span>`)

	for _, block := range modelBlocks {
		// Extract model information from the block
		titleMatch := titleRegex.FindStringSubmatch(block)
		descMatch := descRegex.FindStringSubmatch(block)
		sizeMatches := sizeRegex.FindAllStringSubmatch(block, -1)
		pullsMatch := pullsRegex.FindStringSubmatch(block)
		tagsMatch := tagsRegex.FindStringSubmatch(block)
		updatedMatch := updatedRegex.FindStringSubmatch(block)

		if len(titleMatch) < 2 {
			continue // Skip if no title found
		}

		name := strings.TrimSpace(titleMatch[1])
		name = formatModelName(name)

		// Create model with extracted information
		model := Model{
			Name: name,
		}

		if len(descMatch) >= 2 {
			model.Description = strings.TrimSpace(descMatch[1])
		}

		// Collect all sizes for this model
		var sizes []string
		for _, sizeMatch := range sizeMatches {
			if len(sizeMatch) >= 2 {
				size := strings.TrimSpace(sizeMatch[1])
				if size != "" {
					sizes = append(sizes, size)
				}
			}
		}
		// Sort sizes by their numeric value
		sort.Slice(sizes, func(i, j int) bool {
			// Extract numeric values from size strings
			numI := extractNumericValue(sizes[i])
			numJ := extractNumericValue(sizes[j])
			return numI < numJ
		})
		model.Size = strings.Join(sizes, ", ")

		if len(pullsMatch) >= 2 {
			model.Pulls = strings.TrimSpace(pullsMatch[1])
		}

		if len(tagsMatch) >= 2 {
			model.Tags = strings.TrimSpace(tagsMatch[1])
		}

		if len(updatedMatch) >= 2 {
			model.Updated = strings.TrimSpace(updatedMatch[1])
		}

		models = append(models, model)
	}

	// Sort models by update time before returning
	sortModelsByUpdateTime(models)
	return models, nil
}

// sortModelsByUpdateTime sorts models by their update time, most recent first
func sortModelsByUpdateTime(models []Model) {
	sort.Slice(models, func(i, j int) bool {
		timeI := parseUpdateTime(models[i].Updated)
		timeJ := parseUpdateTime(models[j].Updated)
		return timeI.After(timeJ)
	})
}

// parseUpdateTime parses the update time string into a time.Time
func parseUpdateTime(updated string) time.Time {
	if updated == "" {
		return time.Time{} // Return zero time for empty strings
	}

	// Common time formats used in the update field
	formats := []string{
		"2006-01-02",                // YYYY-MM-DD
		"Jan 2, 2006",               // MMM D, YYYY
		"January 2, 2006",           // MMMM D, YYYY
		"2 Jan 2006",                // D MMM YYYY
		"2006-01-02 15:04:05 -0700", // Full timestamp with timezone
		"2006-01-02T15:04:05-07:00", // ISO format
	}

	// Try to parse with each format
	for _, format := range formats {
		if t, err := time.Parse(format, updated); err == nil {
			return t
		}
	}

	// If we can't parse the exact date, try to parse relative times
	lower := strings.ToLower(updated)
	if strings.HasSuffix(lower, " ago") {
		duration := strings.TrimSuffix(lower, " ago")
		now := time.Now()

		// Parse different duration formats
		switch {
		case strings.HasSuffix(duration, " minutes"):
			if mins, err := strconv.Atoi(strings.TrimSuffix(duration, " minutes")); err == nil {
				return now.Add(-time.Duration(mins) * time.Minute)
			}
		case strings.HasSuffix(duration, " hours"):
			if hours, err := strconv.Atoi(strings.TrimSuffix(duration, " hours")); err == nil {
				return now.Add(-time.Duration(hours) * time.Hour)
			}
		case strings.HasSuffix(duration, " days"):
			if days, err := strconv.Atoi(strings.TrimSuffix(duration, " days")); err == nil {
				return now.AddDate(0, 0, -days)
			}
		case strings.HasSuffix(duration, " weeks"):
			if weeks, err := strconv.Atoi(strings.TrimSuffix(duration, " weeks")); err == nil {
				return now.AddDate(0, 0, -weeks*7)
			}
		case strings.HasSuffix(duration, " months"):
			if months, err := strconv.Atoi(strings.TrimSuffix(duration, " months")); err == nil {
				return now.AddDate(0, -months, 0)
			}
		case strings.HasSuffix(duration, " years"):
			if years, err := strconv.Atoi(strings.TrimSuffix(duration, " years")); err == nil {
				return now.AddDate(-years, 0, 0)
			}
		}
	}
	// yesterday case
	if strings.Contains(lower, "yesterday") {
		return time.Now().AddDate(0, 0, -1)
	}

	return time.Time{} // Return zero time if we can't parse the format
}

// formatModelName formats the model name to match the format used by Ollama
func formatModelName(name string) string {
	// Remove "Model:" prefix if present
	name = strings.TrimPrefix(name, "Model:")
	return strings.TrimSpace(name)
}

// FilterByName filters models by name
func FilterByName(models []Model, filterName string) []Model {
	if filterName == "" {
		return models
	}

	filteredModels := []Model{}
	for _, model := range models {
		if strings.Contains(strings.ToLower(model.Name), strings.ToLower(filterName)) {
			filteredModels = append(filteredModels, model)
		}
	}
	return filteredModels
}

// FilterBySize filters models by their maximum size
// maxSize is the maximum size in billions (e.g., 7 for 7B models)
// If maxSize is <= 0, no filtering is applied
func FilterBySize(models []Model, maxSize float64) []Model {
	if maxSize <= 0 {
		return models
	}

	filteredModels := []Model{}
	for _, model := range models {
		// Split the size string which may contain multiple sizes
		sizes := strings.Split(model.Size, ", ")

		// Check if any size is less than or equal to maxSize
		for _, sizeStr := range sizes {
			size := extractNumericValue(sizeStr)
			if size <= maxSize {
				filteredModels = append(filteredModels, model)
				break
			}
		}
	}
	return filteredModels
}

// extractNumericValue extracts the numeric value from a size string (e.g., "1.5b" -> 1.5)
func extractNumericValue(size string) float64 {
	// Remove the 'b' or 'B' suffix
	size = strings.TrimSuffix(strings.TrimSuffix(size, "b"), "B")
	// Convert to float
	val, _ := strconv.ParseFloat(size, 64)
	return val
}
