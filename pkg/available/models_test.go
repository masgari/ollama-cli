package available

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func sampleModelHTML(name, desc, size, pulls, tags, updated string) string {
	sizeSpans := ""
	for _, s := range strings.Split(size, ", ") {
		if s == "" {
			continue
		}
		sizeSpans += `<span class="inline-flex my-1 items-center rounded-md bg-[#ddf4ff] px-2 py-[2px] text-xs font-medium text-blue-600 sm:text-[13px]">` + s + `</span>`
	}

	return `
<li class="flex items-baseline border-b border-neutral-200 py-6">
  <a href="/library/` + name + `" class="group w-full">
    <div class="flex flex-col mb-1" title="` + name + `">
      <h2 class="truncate text-xl font-medium">
        <span>` + name + `</span>
      </h2>
      <p class="max-w-lg break-words text-neutral-800 text-md">` + desc + `</p>
    </div>
    <div class="flex flex-col">
      <div class="flex flex-wrap space-x-2">
        <span class="inline-flex my-1 items-center rounded-md bg-indigo-50 px-2 py-[2px] text-xs font-medium text-indigo-600 sm:text-[13px]">vision</span>
        ` + sizeSpans + `
      </div>
      <p class="my-1 flex space-x-5 text-[13px] font-medium text-neutral-500">
        <span class="flex items-center">
          <span>` + pulls + `</span>
          <span class="hidden sm:flex">&nbsp;Pulls</span>
        </span>
        <span class="flex items-center">
          <span>` + tags + `</span>
          <span class="hidden sm:flex">&nbsp;Tags</span>
        </span>
        <span class="flex items-center" title="Aug 19, 2026 6:06 PM UTC">
          <span class="hidden sm:flex">Updated&nbsp;</span>
          <span>` + updated + `</span>
        </span>
      </p>
    </div>
  </a>
</li>`
}

func TestFilterByName(t *testing.T) {
	models := []Model{
		{Name: "llama2", Description: "Llama 2 model"},
		{Name: "mistral", Description: "Mistral model"},
		{Name: "llama3", Description: "Llama 3 model"},
	}

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
	html := `<ul role="list" class="grid grid-cols-1">` +
		sampleModelHTML("llama2", "Llama 2 model", "7.0B", "1M", "10", "1 hour ago") +
		sampleModelHTML("gemma2", "Gemma 2 model", "4.0B", "500K", "5", "yesterday") +
		sampleModelHTML("mistral", "Mistral model", "7.0B", "500K", "5", "2 days ago") +
		`
<li
  hx-get="/search?page=2"
  hx-trigger="revealed"
  hx-swap="outerHTML"
  hx-target="this"
></li>
</ul>`

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

	models, err := parseModels(html)
	if err != nil {
		t.Fatalf("parseModels() error = %v", err)
	}

	sort.Slice(models, func(i, j int) bool {
		return models[i].Name < models[j].Name
	})
	sort.Slice(expected, func(i, j int) bool {
		return expected[i].Name < expected[j].Name
	})

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

func TestParseModelsMultipleSizes(t *testing.T) {
	html := `<ul>` + sampleModelHTML("ornith", "Self-improving model", "9b, 35b", "12K", "3", "1 week ago") + `</ul>`

	models, err := parseModels(html)
	if err != nil {
		t.Fatalf("parseModels() error = %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("parseModels() returned %d models, want 1", len(models))
	}
	if models[0].Size != "9b, 35b" {
		t.Errorf("Size = %q, want %q", models[0].Size, "9b, 35b")
	}
	if models[0].Updated != "1 week ago" {
		t.Errorf("Updated = %q, want %q", models[0].Updated, "1 week ago")
	}
}

func TestFetchModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.Header.Get("User-Agent") != "ollama-cli" {
			t.Errorf("Expected User-Agent: ollama-cli, got %s", r.Header.Get("User-Agent"))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<ul>` + sampleModelHTML("llama2", "Llama 2 model", "7.0B", "1M", "10", "1 day ago") + `</ul>`))
	}))
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	fetcher := NewModelFetcher(client, server.URL)

	ctx := context.Background()
	models, err := fetcher.FetchModels(ctx)
	if err != nil {
		t.Fatalf("FetchModels() error = %v", err)
	}

	if len(models) != 1 {
		t.Errorf("FetchModels() returned %d models, want 1", len(models))
	}

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

func TestFetchModelsPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "ollama-cli" {
			t.Errorf("Expected User-Agent: ollama-cli, got %s", r.Header.Get("User-Agent"))
		}

		page := r.URL.Query().Get("page")
		switch page {
		case "", "1":
			if r.Header.Get("HX-Request") != "" {
				t.Errorf("page 1 should not send HX-Request, got %q", r.Header.Get("HX-Request"))
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<!doctype html><ul>` +
				sampleModelHTML("model-a", "First page", "7b", "1M", "2", "1 hour ago") +
				`<li hx-get="/search?page=2" hx-trigger="revealed" hx-swap="outerHTML" hx-target="this"></li></ul>`))
		case "2":
			if r.Header.Get("HX-Request") != "true" {
				t.Errorf("page 2 expected HX-Request: true, got %q", r.Header.Get("HX-Request"))
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(sampleModelHTML("model-b", "Second page", "3b", "500K", "1", "2 days ago") +
				`<li hx-get="/search?page=3" hx-trigger="revealed"></li>`))
		case "3":
			if r.Header.Get("HX-Request") != "true" {
				t.Errorf("page 3 expected HX-Request: true, got %q", r.Header.Get("HX-Request"))
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(sampleModelHTML("model-c", "Third page", "1b", "100K", "1", "1 week ago")))
		default:
			t.Errorf("unexpected page %q", page)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	fetcher := NewModelFetcher(client, server.URL+"?o=newest")

	models, err := fetcher.FetchModels(context.Background())
	if err != nil {
		t.Fatalf("FetchModels() error = %v", err)
	}

	if len(models) != 3 {
		t.Fatalf("FetchModels() returned %d models, want 3: %+v", len(models), models)
	}

	names := map[string]bool{}
	for _, m := range models {
		names[m.Name] = true
	}
	for _, want := range []string{"model-a", "model-b", "model-c"} {
		if !names[want] {
			t.Errorf("missing model %q in %+v", want, models)
		}
	}

	// Newest-first sort: model-a (1 hour) before model-b (2 days) before model-c (1 week)
	if models[0].Name != "model-a" {
		t.Errorf("first model = %q, want model-a", models[0].Name)
	}
}

func TestParseUpdateTimeSingular(t *testing.T) {
	now := time.Now()
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"1 hour ago", time.Hour},
		{"1 day ago", 24 * time.Hour},
		{"1 week ago", 7 * 24 * time.Hour},
	}
	for _, tt := range cases {
		got := parseUpdateTime(tt.in)
		diff := now.Sub(got)
		if diff < tt.want-time.Minute || diff > tt.want+time.Minute {
			t.Errorf("parseUpdateTime(%q) delta = %v, want about %v", tt.in, diff, tt.want)
		}
	}
}

func TestFilterBySize(t *testing.T) {
	models := []Model{
		{Name: "llama2", Size: "7.0B"},
		{Name: "mistral", Size: "14.0B"},
		{Name: "llama3", Size: "3.5B, 7.0B"},
		{Name: "gemma2", Size: "4.0B"},
	}

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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterBySize(models, tt.maxSize)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterBySize() = %v, want %v", got, tt.want)
			}
		})
	}
}
