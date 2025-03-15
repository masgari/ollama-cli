//go:build integration
// +build integration

package available

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestIntegrationFetchModelsFromOllama tests fetching models from the actual ollama.com website
// This test is only run when the integration build tag is specified
func TestIntegrationFetchModelsFromOllama(t *testing.T) {
	// Create a real HTTP client with a reasonable timeout
	client := &http.Client{Timeout: 10 * time.Second}
	fetcher := NewModelFetcher(client, "https://ollama.com/search")

	// Fetch models from the actual ollama.com website
	ctx := context.Background()
	models, err := fetcher.FetchModels(ctx)

	// Check for errors
	if err != nil {
		t.Fatalf("FetchModels() error = %v", err)
	}

	// Verify that we got some models
	if len(models) == 0 {
		t.Fatalf("FetchModels() returned 0 models, expected at least some models")
	}

	// Verify that the models have the expected fields
	for i, model := range models {
		// Check that essential fields are present
		if model.Name == "" {
			t.Errorf("model[%d].Name is empty", i)
		}
	}

	// Check that we have at least some models with descriptions
	hasDescription := false
	for _, m := range models {
		if m.Description != "" {
			hasDescription = true
			break
		}
	}
	if !hasDescription {
		t.Errorf("None of the models have descriptions, expected at least some to have descriptions")
	}

	// Check that we have at least some models with size information
	hasSize := false
	for _, m := range models {
		if m.Size != "" {
			hasSize = true
			break
		}
	}
	if !hasSize {
		t.Errorf("None of the models have size information, expected at least some to have sizes")
	}

	// Check that we have at least some models with pull counts
	hasPulls := false
	for _, m := range models {
		if m.Pulls != "" {
			hasPulls = true
			break
		}
	}
	if !hasPulls {
		t.Errorf("None of the models have pull counts, expected at least some to have pull counts")
	}

	// Check that we have at least some models with tag counts
	hasTags := false
	for _, m := range models {
		if m.Tags != "" {
			hasTags = true
			break
		}
	}
	if !hasTags {
		t.Errorf("None of the models have tag counts, expected at least some to have tag counts")
	}

	// Check that we have at least some models with update times
	hasUpdated := false
	for _, m := range models {
		if m.Updated != "" {
			hasUpdated = true
			break
		}
	}
	if !hasUpdated {
		t.Errorf("None of the models have update times, expected at least some to have update times")
	}

	// Find a model that has all fields populated to verify our regex patterns
	var completeModel *Model
	for i, m := range models {
		if m.Name != "" && m.Description != "" && m.Size != "" &&
			m.Pulls != "" && m.Tags != "" && m.Updated != "" {
			completeModel = &models[i]
			break
		}
	}

	if completeModel != nil {
		t.Logf("Found a complete model: %+v", *completeModel)
	} else {
		t.Logf("Could not find a model with all fields populated")
	}

	// Verify that we can filter models by name
	// Pick a common model name prefix that should exist
	commonPrefix := "llama"
	filteredModels := FilterByName(models, commonPrefix)

	// There should be at least one model with this prefix
	if len(filteredModels) == 0 {
		t.Errorf("FilterByName() with prefix '%s' returned 0 models, expected at least one", commonPrefix)
	}

	// All filtered models should contain the prefix in their name
	for i, model := range filteredModels {
		if !strings.Contains(strings.ToLower(model.Name), commonPrefix) {
			t.Errorf("filteredModels[%d].Name = %s, does not contain prefix '%s'", i, model.Name, commonPrefix)
		}
	}

	// Verify that we can filter models by size
	if hasSize {
		// Find a size that exists in the dataset
		var sizeValue float64
		for _, m := range models {
			if m.Size != "" {
				// Extract the first size if there are multiple
				sizes := strings.Split(m.Size, ", ")
				if len(sizes) > 0 {
					sizeStr := sizes[0]
					// Remove the 'B' suffix
					sizeStr = strings.TrimSuffix(strings.TrimSuffix(sizeStr, "B"), "b")
					var err error
					sizeValue, err = strconv.ParseFloat(sizeStr, 64)
					if err == nil {
						break
					}
				}
			}
		}

		if sizeValue > 0 {
			// Filter by this size
			sizeFilteredModels := FilterBySize(models, sizeValue)

			// There should be at least one model with this size
			if len(sizeFilteredModels) == 0 {
				t.Errorf("FilterBySize() with size <= %f returned 0 models, expected at least one", sizeValue)
			}

			t.Logf("Successfully filtered models by size <= %f, found %d models", sizeValue, len(sizeFilteredModels))
		}
	}

	// Log some information about the models for debugging
	t.Logf("Successfully fetched %d models from ollama.com", len(models))
	if len(models) > 0 {
		t.Logf("First model: %+v", models[0])
	}
}
