package available

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultSearchURL lists models on ollama.com sorted by newest.
	DefaultSearchURL = "https://ollama.com/search?o=newest"
	maxSearchPages   = 50
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

var (
	modelBlockRegex = regexp.MustCompile(`(?s)<li\s+class="flex items-baseline[^"]*".*?</li>`)
	nameRegex       = regexp.MustCompile(`href="/library/([^"]+)"`)
	descRegex       = regexp.MustCompile(`<p class="max-w-lg break-words[^>]*>(.*?)</p>`)
	sizeRegex       = regexp.MustCompile(`<span[^>]*>\s*(\d+(?:\.\d+)?[bB])\s*</span>`)
	pullsRegex      = regexp.MustCompile(`(?s)<span[^>]*>\s*([^<]+?)\s*</span>\s*<span[^>]*>\s*(?:&nbsp;)?\s*Pulls\s*</span>`)
	tagsRegex       = regexp.MustCompile(`(?s)<span[^>]*>\s*([^<]+?)\s*</span>\s*<span[^>]*>\s*(?:&nbsp;)?\s*Tags?\s*</span>`)
	updatedRegex    = regexp.MustCompile(`(?s)Updated(?:&nbsp;|\s)*</span>\s*<span[^>]*>\s*([^<]+?)\s*</span>`)
	updatedTitleRe  = regexp.MustCompile(`title="([^"]+)"[^>]*>\s*(?:<svg[\s\S]*?</svg>\s*)?(?:<span[^>]*>\s*Updated(?:&nbsp;|\s)*</span>\s*)?<span[^>]*>\s*([^<]+?)\s*</span>`)
	nextPageRegex   = regexp.MustCompile(`hx-get="/search\?page=(\d+)"`)
)

// FetchModels fetches the list of available models from the specified URL,
// following HTMX infinite-scroll pagination when present.
func (mf *ModelFetcher) FetchModels(ctx context.Context) ([]Model, error) {
	baseURL, err := url.Parse(mf.url)
	if err != nil {
		return nil, fmt.Errorf("invalid fetch URL: %w", err)
	}

	var allModels []Model
	seen := make(map[string]struct{})
	pageURL := mf.url
	htmxRequest := false

	for page := 0; page < maxSearchPages; page++ {
		body, err := mf.fetchPage(ctx, pageURL, htmxRequest)
		if err != nil {
			return nil, err
		}

		models, err := parseModels(body)
		if err != nil {
			// First page with no models is a hard error; later empty pages end pagination.
			if page == 0 {
				return nil, fmt.Errorf("failed to parse response: %w", err)
			}
			break
		}

		for _, model := range models {
			if _, ok := seen[model.Name]; ok {
				continue
			}
			seen[model.Name] = struct{}{}
			allModels = append(allModels, model)
		}

		nextPage := findNextPage(body)
		if nextPage == 0 {
			break
		}

		nextURL := *baseURL
		q := nextURL.Query()
		q.Set("page", strconv.Itoa(nextPage))
		nextURL.RawQuery = q.Encode()
		pageURL = nextURL.String()
		htmxRequest = true
	}

	if len(allModels) == 0 {
		return nil, fmt.Errorf("no models found in response")
	}

	sortModelsByUpdateTime(allModels)
	return allModels, nil
}

func (mf *ModelFetcher) fetchPage(ctx context.Context, pageURL string, htmxRequest bool) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "ollama-cli")
	if htmxRequest {
		req.Header.Set("HX-Request", "true")
	}

	resp, err := mf.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

func findNextPage(html string) int {
	match := nextPageRegex.FindStringSubmatch(html)
	if len(match) < 2 {
		return 0
	}
	page, err := strconv.Atoi(match[1])
	if err != nil {
		return 0
	}
	return page
}

// FetchModels fetches models from ollama.com using the default search URL.
func FetchModels(ctx context.Context, timeout int) ([]Model, error) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	fetcher := NewModelFetcher(client, DefaultSearchURL)
	return fetcher.FetchModels(ctx)
}

// parseModels parses the HTML response from ollama.com/search
func parseModels(html string) ([]Model, error) {
	modelBlocks := modelBlockRegex.FindAllString(html, -1)
	if len(modelBlocks) == 0 {
		return nil, fmt.Errorf("no models found in response")
	}

	var models []Model
	for _, block := range modelBlocks {
		nameMatch := nameRegex.FindStringSubmatch(block)
		if len(nameMatch) < 2 {
			continue // Skip HTMX sentinels and other non-model list items
		}

		name := formatModelName(strings.TrimSpace(nameMatch[1]))
		model := Model{Name: name}

		if descMatch := descRegex.FindStringSubmatch(block); len(descMatch) >= 2 {
			model.Description = strings.TrimSpace(descMatch[1])
		}

		var sizes []string
		for _, sizeMatch := range sizeRegex.FindAllStringSubmatch(block, -1) {
			if len(sizeMatch) >= 2 {
				size := strings.TrimSpace(sizeMatch[1])
				if size != "" {
					sizes = append(sizes, size)
				}
			}
		}
		sort.Slice(sizes, func(i, j int) bool {
			return extractNumericValue(sizes[i]) < extractNumericValue(sizes[j])
		})
		model.Size = strings.Join(sizes, ", ")

		if pullsMatch := pullsRegex.FindStringSubmatch(block); len(pullsMatch) >= 2 {
			model.Pulls = strings.TrimSpace(pullsMatch[1])
		}

		if tagsMatch := tagsRegex.FindStringSubmatch(block); len(tagsMatch) >= 2 {
			model.Tags = strings.TrimSpace(tagsMatch[1])
		}

		if updatedMatch := updatedRegex.FindStringSubmatch(block); len(updatedMatch) >= 2 {
			model.Updated = strings.TrimSpace(updatedMatch[1])
		} else if titleMatch := updatedTitleRe.FindStringSubmatch(block); len(titleMatch) >= 3 {
			// Prefer relative display text; fall back to absolute title if needed.
			model.Updated = strings.TrimSpace(titleMatch[2])
			if model.Updated == "" {
				model.Updated = strings.TrimSpace(titleMatch[1])
			}
		}

		models = append(models, model)
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("no models found in response")
	}

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
		return time.Time{}
	}

	formats := []string{
		"2006-01-02",
		"Jan 2, 2006",
		"January 2, 2006",
		"2 Jan 2006",
		"Jan 2, 2006 3:04 PM MST",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02T15:04:05-07:00",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, updated); err == nil {
			return t
		}
	}

	lower := strings.ToLower(updated)
	if strings.Contains(lower, "yesterday") {
		return time.Now().AddDate(0, 0, -1)
	}

	if strings.HasSuffix(lower, " ago") {
		duration := strings.TrimSpace(strings.TrimSuffix(lower, " ago"))
		now := time.Now()

		unitParsers := []struct {
			singular string
			plural   string
			apply    func(int) time.Time
		}{
			{"minute", "minutes", func(n int) time.Time { return now.Add(-time.Duration(n) * time.Minute) }},
			{"hour", "hours", func(n int) time.Time { return now.Add(-time.Duration(n) * time.Hour) }},
			{"day", "days", func(n int) time.Time { return now.AddDate(0, 0, -n) }},
			{"week", "weeks", func(n int) time.Time { return now.AddDate(0, 0, -n*7) }},
			{"month", "months", func(n int) time.Time { return now.AddDate(0, -n, 0) }},
			{"year", "years", func(n int) time.Time { return now.AddDate(-n, 0, 0) }},
		}

		for _, unit := range unitParsers {
			if strings.HasSuffix(duration, " "+unit.plural) {
				if n, err := strconv.Atoi(strings.TrimSpace(strings.TrimSuffix(duration, " "+unit.plural))); err == nil {
					return unit.apply(n)
				}
			}
			if strings.HasSuffix(duration, " "+unit.singular) {
				if n, err := strconv.Atoi(strings.TrimSpace(strings.TrimSuffix(duration, " "+unit.singular))); err == nil {
					return unit.apply(n)
				}
			}
		}
	}

	return time.Time{}
}

// formatModelName formats the model name to match the format used by Ollama
func formatModelName(name string) string {
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
		sizes := strings.Split(model.Size, ", ")
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
	size = strings.TrimSuffix(strings.TrimSuffix(size, "b"), "B")
	val, _ := strconv.ParseFloat(size, 64)
	return val
}
