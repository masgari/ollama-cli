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

func sampleModelHTML(name, desc, size, pulls string) string {
	sizeSpans := ""
	for _, s := range strings.Split(size, ", ") {
		if s == "" {
			continue
		}
		sizeSpans += `<span class="inline-flex items-center rounded-md bg-[#ddf4ff] px-2 py-0.5 text-xs font-medium text-blue-600 sm:text-[13px]">` + s + `</span>`
	}

	pullsHTML := ""
	if pulls != "" {
		pullsHTML = `<span class="inline-flex shrink-0 items-center gap-1.5 text-[13px] leading-7 tabular-nums text-black/60" title="1,000 downloads"><svg class="h-[1.15em] w-[1.15em] flex-none" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true"></svg><span >` + pulls + `</span></span>`
	}

	return `
<li class="border-b border-black/[0.08]">
  <a href="/library/` + name + `" class="group flex items-start justify-between gap-6 py-6">
    <div class="min-w-0 flex-1">
      <h2 class="truncate text-xl font-medium leading-7 tracking-tight underline-offset-4 group-hover:underline" title="` + name + `">
        <span >` + name + `</span>
      </h2>
      <p class="mt-1 max-w-2xl truncate text-sm leading-6 text-black/60" title="` + desc + `">` + desc + `</p>
      <div class="mt-3 flex flex-wrap items-center gap-2 text-[13px] leading-5 text-black/60">
        <span class="inline-flex items-center rounded-md bg-indigo-50 px-2 py-0.5 text-xs font-medium text-indigo-600 sm:text-[13px]">vision</span>
        ` + sizeSpans + `
      </div>
    </div>
    ` + pullsHTML + `
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

func TestBuildSearchURL(t *testing.T) {
	tests := []struct {
		name    string
		opts    SearchOptions
		want    string
		wantErr bool
	}{
		{
			name: "defaults",
			opts: SearchOptions{},
			want: "https://ollama.com/search?c=local&o=newest",
		},
		{
			name: "popular cloud with vision and max-gb",
			opts: SearchOptions{
				Sort:         "popular",
				Where:        "cloud",
				Capabilities: []string{"vision", "tools"},
				MaxGB:        16,
			},
			want: "https://ollama.com/search?c=cloud&c=vision&c=tools&o=popular&s=16",
		},
		{
			name: "all where omits location filter",
			opts: SearchOptions{Sort: "name", Where: "all"},
			want: "https://ollama.com/search?o=name",
		},
		{
			name:    "invalid sort",
			opts:    SearchOptions{Sort: "stars"},
			wantErr: true,
		},
		{
			name:    "invalid max-gb",
			opts:    SearchOptions{MaxGB: 12},
			wantErr: true,
		},
		{
			name:    "invalid capability",
			opts:    SearchOptions{Capabilities: []string{"audio"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildSearchURL(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BuildSearchURL() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("BuildSearchURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseLibraryFileSize(t *testing.T) {
	html := `<p class="flex text-neutral-500">1.3GB · 256K context window · Text, Image · 4</p>`
	if got := parseLibraryFileSize(html); got != "1.3GB" {
		t.Errorf("parseLibraryFileSize() = %q, want 1.3GB", got)
	}
}

func TestEnrichFileSizes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/library/llama2":
			w.Write([]byte(`<p class="flex text-neutral-500">3.8GB · 4K context window · Text · 1</p>`))
		case "/library/missing":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	oldBase := libraryURLBase
	libraryURLBase = server.URL
	defer func() { libraryURLBase = oldBase }()

	models := []Model{
		{Name: "llama2"},
		{Name: "missing"},
	}
	EnrichFileSizes(context.Background(), server.Client(), models)

	if models[0].FileSize != "3.8GB" {
		t.Errorf("models[0].FileSize = %q, want 3.8GB", models[0].FileSize)
	}
	if models[1].FileSize != "" {
		t.Errorf("models[1].FileSize = %q, want empty", models[1].FileSize)
	}
}

func TestParseModels(t *testing.T) {
	html := `<ul role="list">` +
		sampleModelHTML("llama2", "Llama 2 model", "7.0B", "1M") +
		sampleModelHTML("gemma2", "Gemma 2 model", "4.0B", "500K") +
		sampleModelHTML("mistral", "Mistral model", "7.0B", "500K") +
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
			Name:         "llama2",
			Description:  "Llama 2 model",
			Size:         "7.0B",
			Capabilities: "vision",
			Pulls:        "1M",
		},
		{
			Name:         "gemma2",
			Description:  "Gemma 2 model",
			Size:         "4.0B",
			Capabilities: "vision",
			Pulls:        "500K",
		},
		{
			Name:         "mistral",
			Description:  "Mistral model",
			Size:         "7.0B",
			Capabilities: "vision",
			Pulls:        "500K",
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
		t.Fatalf("parseModels() returned %d models, want %d", len(models), len(expected))
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
		if model.Capabilities != expected[i].Capabilities {
			t.Errorf("model[%d].Capabilities = %s, want %s", i, model.Capabilities, expected[i].Capabilities)
		}
		if model.Tags != "" {
			t.Errorf("model[%d].Tags = %q, want empty", i, model.Tags)
		}
		if model.Updated != "" {
			t.Errorf("model[%d].Updated = %q, want empty", i, model.Updated)
		}
	}
}

func TestParseModelsMultipleSizes(t *testing.T) {
	html := `<ul>` + sampleModelHTML("ornith", "Self-improving model", "9b, 35b", "12K") + `</ul>`

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
}

func TestParseModelsMillionSizes(t *testing.T) {
	html := `<ul>` + sampleModelHTML("embeddinggemma-2", "Embedding model", "270m, 440m, 740m", "24.5K") + `</ul>`

	models, err := parseModels(html)
	if err != nil {
		t.Fatalf("parseModels() error = %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("parseModels() returned %d models, want 1", len(models))
	}
	if models[0].Size != "270m, 440m, 740m" {
		t.Errorf("Size = %q, want %q", models[0].Size, "270m, 440m, 740m")
	}
	if models[0].Pulls != "24.5K" {
		t.Errorf("Pulls = %q, want %q", models[0].Pulls, "24.5K")
	}
}

func TestParseModelsCloudCard(t *testing.T) {
	html := `<ul>` + sampleModelHTML("mistral-large-4", "Mistral&#39;s open-weight model", "", "") + `</ul>`

	models, err := parseModels(html)
	if err != nil {
		t.Fatalf("parseModels() error = %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("parseModels() returned %d models, want 1", len(models))
	}
	if models[0].Name != "mistral-large-4" {
		t.Errorf("Name = %q, want mistral-large-4", models[0].Name)
	}
	if models[0].Description != "Mistral's open-weight model" {
		t.Errorf("Description = %q, want unescaped apostrophe", models[0].Description)
	}
	if models[0].Size != "" {
		t.Errorf("Size = %q, want empty", models[0].Size)
	}
	if models[0].Pulls != "" {
		t.Errorf("Pulls = %q, want empty", models[0].Pulls)
	}
}

func TestParseModelsPreservesOrderWithoutUpdated(t *testing.T) {
	html := `<ul>` +
		sampleModelHTML("first", "First", "7b", "1M") +
		sampleModelHTML("second", "Second", "3b", "500K") +
		sampleModelHTML("third", "Third", "1b", "100K") +
		`</ul>`

	models, err := parseModels(html)
	if err != nil {
		t.Fatalf("parseModels() error = %v", err)
	}
	want := []string{"first", "second", "third"}
	for i, name := range want {
		if models[i].Name != name {
			t.Errorf("models[%d].Name = %q, want %q", i, models[i].Name, name)
		}
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
		w.Write([]byte(`<ul>` + sampleModelHTML("llama2", "Llama 2 model", "7.0B", "1M") + `</ul>`))
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
				sampleModelHTML("model-a", "First page", "7b", "1M") +
				`<li hx-get="/search?page=2" hx-trigger="revealed" hx-swap="outerHTML" hx-target="this"></li></ul>`))
		case "2":
			if r.Header.Get("HX-Request") != "true" {
				t.Errorf("page 2 expected HX-Request: true, got %q", r.Header.Get("HX-Request"))
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(sampleModelHTML("model-b", "Second page", "3b", "500K") +
				`<li hx-get="/search?page=3" hx-trigger="revealed"></li>`))
		case "3":
			if r.Header.Get("HX-Request") != "true" {
				t.Errorf("page 3 expected HX-Request: true, got %q", r.Header.Get("HX-Request"))
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(sampleModelHTML("model-c", "Third page", "1b", "100K")))
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

	// Server order preserved when Updated is absent
	if models[0].Name != "model-a" {
		t.Errorf("first model = %q, want model-a", models[0].Name)
	}
	if models[1].Name != "model-b" {
		t.Errorf("second model = %q, want model-b", models[1].Name)
	}
	if models[2].Name != "model-c" {
		t.Errorf("third model = %q, want model-c", models[2].Name)
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
		{Name: "embedding", Size: "270m, 740m"},
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
				{Name: "embedding", Size: "270m, 740m"},
			},
		},
		{
			name:    "Filter models with size <= 4B",
			maxSize: 4,
			want: []Model{
				{Name: "llama3", Size: "3.5B, 7.0B"},
				{Name: "gemma2", Size: "4.0B"},
				{Name: "embedding", Size: "270m, 740m"},
			},
		},
		{
			name:    "Filter models with size <= 3B",
			maxSize: 3,
			want: []Model{
				{Name: "embedding", Size: "270m, 740m"},
			},
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
