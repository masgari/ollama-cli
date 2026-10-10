package available

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	searchBaseURL  = "https://ollama.com/search"
	maxSearchPages = 50
	enrichWorkers  = 8
)

// libraryURLBase is the origin used when enriching file sizes from library pages.
var libraryURLBase = "https://ollama.com"

// DefaultSearchURL lists local models on ollama.com sorted by newest.
var DefaultSearchURL = mustBuildSearchURL(SearchOptions{
	Sort:  "newest",
	Where: "local",
})

// SearchOptions controls ollama.com/search query parameters.
type SearchOptions struct {
	Sort         string   // newest, name, popular
	Where        string   // local, cloud, all
	Capabilities []string // tools, thinking, vision, embedding, decision
	MaxGB        int      // 0 (any), 8, 16, 32, 64
}

// Model represents a model available on ollama.com
type Model struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Size         string `json:"size,omitempty"`
	FileSize     string `json:"file_size,omitempty"`
	Capabilities string `json:"capabilities,omitempty"`
	Pulls        string `json:"pulls,omitempty"`
	Tags         string `json:"tags,omitempty"`
	Updated      string `json:"updated,omitempty"`
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
	modelBlockRegex   = regexp.MustCompile(`(?s)<li\b[^>]*>\s*<a href="/library/[^"]+".*?</li>`)
	nameRegex         = regexp.MustCompile(`href="/library/([^"]+)"`)
	descRegex         = regexp.MustCompile(`<p class="[^"]*max-w-[^"]*"[^>]*>(.*?)</p>`)
	sizeRegex         = regexp.MustCompile(`bg-\[#ddf4ff\][^>]*>\s*(\d+(?:\.\d+)?[bBmM])\s*<`)
	capabilityRegex   = regexp.MustCompile(`bg-indigo-50[^>]*>\s*([a-zA-Z]+)\s*<`)
	pullsRegex        = regexp.MustCompile(`(?s)title="[\d,]+ downloads"[^>]*>.*?<span[^>]*>\s*([^<]+?)\s*</span>`)
	pullsTitleRegex   = regexp.MustCompile(`title="([\d,]+) downloads"`)
	tagsRegex         = regexp.MustCompile(`(?s)<span[^>]*>\s*([^<]+?)\s*</span>\s*<span[^>]*>\s*(?:&nbsp;)?\s*Tags?\s*</span>`)
	updatedRegex      = regexp.MustCompile(`(?s)Updated(?:&nbsp;|\s)*</span>\s*<span[^>]*>\s*([^<]+?)\s*</span>`)
	updatedTitleRe    = regexp.MustCompile(`title="([^"]+)"[^>]*>\s*(?:<svg[\s\S]*?</svg>\s*)?<span[^>]*>\s*Updated(?:&nbsp;|\s)*</span>\s*<span[^>]*>\s*([^<]+?)\s*</span>`)
	nextPageRegex     = regexp.MustCompile(`hx-get="/search\?page=(\d+)"`)
	libraryFileSizeRe = regexp.MustCompile(`(\d+(?:\.\d+)?\s*[GMK]B)\s*·`)

	validSorts = map[string]bool{
		"newest":  true,
		"name":    true,
		"popular": true,
	}
	validWhere = map[string]bool{
		"local": true,
		"cloud": true,
		"all":   true,
	}
	validCapabilities = map[string]bool{
		"tools":     true,
		"thinking":  true,
		"vision":    true,
		"embedding": true,
		"decision":  true,
	}
	validMaxGB = map[int]bool{
		0:  true,
		8:  true,
		16: true,
		32: true,
		64: true,
	}
)

func mustBuildSearchURL(opts SearchOptions) string {
	u, err := BuildSearchURL(opts)
	if err != nil {
		panic(err)
	}
	return u
}

// BuildSearchURL builds an ollama.com/search URL from the given options.
func BuildSearchURL(opts SearchOptions) (string, error) {
	if opts.Sort == "" {
		opts.Sort = "newest"
	}
	if opts.Where == "" {
		opts.Where = "local"
	}

	sortKey := strings.ToLower(opts.Sort)
	if !validSorts[sortKey] {
		return "", fmt.Errorf("invalid sort %q (want newest, name, or popular)", opts.Sort)
	}

	whereKey := strings.ToLower(opts.Where)
	if !validWhere[whereKey] {
		return "", fmt.Errorf("invalid where %q (want local, cloud, or all)", opts.Where)
	}

	if !validMaxGB[opts.MaxGB] {
		return "", fmt.Errorf("invalid max-gb %d (want 0, 8, 16, 32, or 64)", opts.MaxGB)
	}

	q := url.Values{}
	q.Set("o", sortKey)

	if whereKey == "local" || whereKey == "cloud" {
		q.Add("c", whereKey)
	}

	seenCaps := make(map[string]bool)
	for _, cap := range opts.Capabilities {
		capKey := strings.ToLower(strings.TrimSpace(cap))
		if capKey == "" {
			continue
		}
		if !validCapabilities[capKey] {
			return "", fmt.Errorf("invalid capability %q (want tools, thinking, vision, embedding, or decision)", cap)
		}
		if seenCaps[capKey] {
			continue
		}
		seenCaps[capKey] = true
		q.Add("c", capKey)
	}

	if opts.MaxGB > 0 {
		q.Set("s", strconv.Itoa(opts.MaxGB))
	}

	return searchBaseURL + "?" + q.Encode(), nil
}

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

	maybeSortModelsByUpdateTime(allModels)
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

// EnrichFileSizes fills FileSize for each model by fetching its library page.
// Failures are best-effort: models without a parseable size are left empty.
func EnrichFileSizes(ctx context.Context, client *http.Client, models []Model) {
	if len(models) == 0 {
		return
	}

	workers := enrichWorkers
	if len(models) < workers {
		workers = len(models)
	}

	jobs := make(chan int, len(models))
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				if ctx.Err() != nil {
					return
				}
				size, err := fetchLibraryFileSize(ctx, client, models[i].Name)
				if err == nil && size != "" {
					models[i].FileSize = size
				}
			}
		}()
	}

	for i := range models {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
}

func fetchLibraryFileSize(ctx context.Context, client *http.Client, name string) (string, error) {
	pageURL := libraryURLBase + "/library/" + url.PathEscape(name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "ollama-cli")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return parseLibraryFileSize(string(body)), nil
}

func parseLibraryFileSize(pageHTML string) string {
	match := libraryFileSizeRe.FindStringSubmatch(pageHTML)
	if len(match) < 2 {
		return ""
	}
	return strings.ReplaceAll(strings.TrimSpace(match[1]), " ", "")
}

// parseModels parses the HTML response from ollama.com/search
func parseModels(pageHTML string) ([]Model, error) {
	modelBlocks := modelBlockRegex.FindAllString(pageHTML, -1)
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
			model.Description = strings.TrimSpace(html.UnescapeString(descMatch[1]))
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

		var caps []string
		for _, capMatch := range capabilityRegex.FindAllStringSubmatch(block, -1) {
			if len(capMatch) >= 2 {
				cap := strings.TrimSpace(capMatch[1])
				if cap != "" {
					caps = append(caps, cap)
				}
			}
		}
		model.Capabilities = strings.Join(caps, ", ")

		if pullsMatch := pullsRegex.FindStringSubmatch(block); len(pullsMatch) >= 2 {
			model.Pulls = strings.TrimSpace(pullsMatch[1])
		} else if titleMatch := pullsTitleRegex.FindStringSubmatch(block); len(titleMatch) >= 2 {
			model.Pulls = strings.TrimSpace(titleMatch[1])
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

	maybeSortModelsByUpdateTime(models)
	return models, nil
}

// maybeSortModelsByUpdateTime sorts by update time when present; otherwise
// preserves server order (e.g. newest-first from ?o=newest).
func maybeSortModelsByUpdateTime(models []Model) {
	for _, m := range models {
		if m.Updated != "" {
			sortModelsByUpdateTime(models)
			return
		}
	}
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

// extractNumericValue extracts the numeric value from a size string in billions
// (e.g., "1.5b" -> 1.5, "270m" -> 0.27).
func extractNumericValue(size string) float64 {
	size = strings.TrimSpace(size)
	lower := strings.ToLower(size)
	switch {
	case strings.HasSuffix(lower, "m"):
		val, _ := strconv.ParseFloat(strings.TrimSuffix(lower, "m"), 64)
		return val / 1000
	case strings.HasSuffix(lower, "b"):
		val, _ := strconv.ParseFloat(strings.TrimSuffix(lower, "b"), 64)
		return val
	default:
		val, _ := strconv.ParseFloat(lower, 64)
		return val
	}
}
