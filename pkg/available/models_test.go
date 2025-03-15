package available

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestFilterByName(t *testing.T) {
	// Test data
	models := []Model{
		{Name: "llama2", Description: "Llama 2 model"},
		{Name: "mistral", Description: "Mistral model"},
		{Name: "llama3", Description: "Llama 3 model"},
	}

	// Test cases
	tests := []struct {
		name       string
		filterName string
		want       []Model
	}{
		{
			name:       "Empty filter returns all models",
			filterName: "",
			want:       models,
		},
		{
			name:       "Filter by llama returns llama models",
			filterName: "llama",
			want: []Model{
				{Name: "llama2", Description: "Llama 2 model"},
				{Name: "llama3", Description: "Llama 3 model"},
			},
		},
		{
			name:       "Filter is case insensitive",
			filterName: "MISTRAL",
			want: []Model{
				{Name: "mistral", Description: "Mistral model"},
			},
		},
		{
			name:       "Non-matching filter returns empty slice",
			filterName: "nonexistent",
			want:       []Model{},
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterByName(models, tt.filterName)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterByName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseModels(t *testing.T) {
	// Sample HTML response
	html := `
	<ul>
		<li x-test-model>
			<span x-test-search-response-title>llama2</span>
			<p class="max-w-lg break-words">Llama 2 model</p>
			<span x-test-size>7.0B</span>
			<span x-test-pull-count>1M</span>
			<span x-test-tag-count>10</span>
			<span x-test-updated>1 hour ago</span>
		</li>
		<li x-test-model>
			<span x-test-search-response-title>gemma2</span>
			<p class="max-w-lg break-words">Gemma 2 model</p>
			<span x-test-size>4.0B</span>
			<span x-test-pull-count>500K</span>
			<span x-test-tag-count>5</span>
			<span x-test-updated>yesterday</span>
		</li>
		<li x-test-model>
			<span x-test-search-response-title>mistral</span>
			<p class="max-w-lg break-words">Mistral model</p>
			<span x-test-size>7.0B</span>
			<span x-test-pull-count>500K</span>
			<span x-test-tag-count>5</span>
			<span x-test-updated>2 days ago</span>
		</li>
	</ul>
	`

	// Expected models
	expected := []Model{
		{
			Name:        "llama2",
			Description: "Llama 2 model",
			Size:        "7.0B",
			Pulls:       "1M",
			Tags:        "10",
			Updated:     "1 hour ago",
		},
		{
			Name:        "gemma2",
			Description: "Gemma 2 model",
			Size:        "4.0B",
			Pulls:       "500K",
			Tags:        "5",
			Updated:     "yesterday",
		},
		{
			Name:        "mistral",
			Description: "Mistral model",
			Size:        "7.0B",
			Pulls:       "500K",
			Tags:        "5",
			Updated:     "2 days ago",
		},
	}

	// Parse models
	models, err := parseModels(html)
	if err != nil {
		t.Fatalf("parseModels() error = %v", err)
	}

	// Sort both slices by name for consistent comparison
	sort.Slice(models, func(i, j int) bool {
		return models[i].Name < models[j].Name
	})
	sort.Slice(expected, func(i, j int) bool {
		return expected[i].Name < expected[j].Name
	})

	// Compare results
	if len(models) != len(expected) {
		t.Errorf("parseModels() returned %d models, want %d", len(models), len(expected))
	}

	for i, model := range models {
		if model.Name != expected[i].Name {
			t.Errorf("model[%d].Name = %s, want %s", i, model.Name, expected[i].Name)
		}
		if model.Description != expected[i].Description {
			t.Errorf("model[%d].Description = %s, want %s", i, model.Description, expected[i].Description)
		}
		if model.Size != expected[i].Size {
			t.Errorf("model[%d].Size = %s, want %s", i, model.Size, expected[i].Size)
		}
		if model.Pulls != expected[i].Pulls {
			t.Errorf("model[%d].Pulls = %s, want %s", i, model.Pulls, expected[i].Pulls)
		}
		if model.Tags != expected[i].Tags {
			t.Errorf("model[%d].Tags = %s, want %s", i, model.Tags, expected[i].Tags)
		}
		if model.Updated != expected[i].Updated {
			t.Errorf("model[%d].Updated = %s, want %s", i, model.Updated, expected[i].Updated)
		}
	}
}

func TestFetchModels(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		// We're not checking the path anymore since we'll use the full URL
		// if r.URL.Path != "/search" {
		// 	t.Errorf("Expected /search path, got %s", r.URL.Path)
		// }

		// Check User-Agent header
		if r.Header.Get("User-Agent") != "ollama-cli" {
			t.Errorf("Expected User-Agent: ollama-cli, got %s", r.Header.Get("User-Agent"))
		}

		// Return a sample response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
		<ul>
			<li x-test-model>
				<span x-test-search-response-title>llama2</span>
				<p class="max-w-lg break-words">Llama 2 model</p>
				<span x-test-size>7.0B</span>
				<span x-test-pull-count>1M</span>
				<span x-test-tag-count>10</span>
				<span x-test-updated>1 day ago</span>
			</li>
		</ul>
		`))
	}))
	defer server.Close()

	// Create a custom ModelFetcher that uses the test server
	client := &http.Client{Timeout: 5 * time.Second}
	fetcher := NewModelFetcher(client, server.URL)

	// Fetch models using the custom fetcher
	ctx := context.Background()
	models, err := fetcher.FetchModels(ctx)

	// Check for errors
	if err != nil {
		t.Fatalf("FetchModels() error = %v", err)
	}

	// Check the number of models returned
	if len(models) != 1 {
		t.Errorf("FetchModels() returned %d models, want 1", len(models))
	}

	// Check the model details
	model := models[0]
	if model.Name != "llama2" {
		t.Errorf("model.Name = %s, want llama2", model.Name)
	}
	if model.Description != "Llama 2 model" {
		t.Errorf("model.Description = %s, want Llama 2 model", model.Description)
	}
	if model.Size != "7.0B" {
		t.Errorf("model.Size = %s, want 7.0B", model.Size)
	}
	if model.Pulls != "1M" {
		t.Errorf("model.Pulls = %s, want 1M", model.Pulls)
	}
	if model.Tags != "10" {
		t.Errorf("model.Tags = %s, want 10", model.Tags)
	}
	if model.Updated != "1 day ago" {
		t.Errorf("model.Updated = %s, want 1 day ago", model.Updated)
	}
}

func TestFilterBySize(t *testing.T) {
	// Test data
	models := []Model{
		{Name: "llama2", Size: "7.0B"},
		{Name: "mistral", Size: "14.0B"},
		{Name: "llama3", Size: "3.5B, 7.0B"},
		{Name: "gemma2", Size: "4.0B"},
	}

	// Test cases
	tests := []struct {
		name    string
		maxSize float64
		want    []Model
	}{
		{
			name:    "No size limit returns all models",
			maxSize: 0,
			want:    models,
		},
		{
			name:    "Filter models with size <= 7B",
			maxSize: 7,
			want: []Model{
				{Name: "llama2", Size: "7.0B"},
				{Name: "llama3", Size: "3.5B, 7.0B"},
				{Name: "gemma2", Size: "4.0B"},
			},
		},
		{
			name:    "Filter models with size <= 4B",
			maxSize: 4,
			want: []Model{
				{Name: "llama3", Size: "3.5B, 7.0B"},
				{Name: "gemma2", Size: "4.0B"},
			},
		},
		{
			name:    "Filter models with size <= 3B returns none",
			maxSize: 3,
			want:    []Model{},
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterBySize(models, tt.maxSize)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterBySize() = %v, want %v", got, tt.want)
			}
		})
	}
}
