package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/masgari/ollama-cli/pkg/available"
	"github.com/masgari/ollama-cli/pkg/client"
	"github.com/masgari/ollama-cli/pkg/output"
	"github.com/ollama/ollama/api"
	"github.com/spf13/cobra"
)

// outdatedCmd represents the outdated command
var outdatedCmd = &cobra.Command{
	Use:     "outdated",
	Aliases: []string{"out"},
	Short:   "Check for outdated models on the Ollama server",
	Long:    `Check if installed models on the Ollama server are outdated compared to their versions in the Ollama library.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the output format from flags
		outputFormat, _ := cmd.Flags().GetString("output")
		showDetails, _ := cmd.Flags().GetBool("details")

		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()

		// Create Ollama client
		ollamaClient, err := client.New(cfg)
		if err != nil {
			return fmt.Errorf("failed to create Ollama client: %w", err)
		}

		// Fetch installed models
		installedModels, err := ollamaClient.ListModels(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list installed models: %w", err)
		}

		// If no models are installed, print a message and return
		if len(installedModels.Models) == 0 {
			output.Default.InfoPrintln("No models found on the Ollama server.")
			return nil
		}

		// Fetch available models from ollama.com
		availableModels, err := available.FetchModels(ctx, timeout)
		if err != nil {
			return fmt.Errorf("failed to fetch available models: %w", err)
		}

		// Create a map of available models for quick lookup
		availableModelMap := make(map[string]available.Model)
		for _, model := range availableModels {
			// Extract base model name (without tags)
			baseName := strings.Split(model.Name, ":")[0]
			availableModelMap[baseName] = model
		}

		// Check for outdated models
		var outdatedModels []OutdatedModel
		for _, installedModel := range installedModels.Models {
			// Skip models that don't match the filter
			if filterName != "" && !strings.Contains(strings.ToLower(installedModel.Name), strings.ToLower(filterName)) {
				continue
			}

			// Extract base model name (without tags)
			baseName := strings.Split(installedModel.Name, ":")[0]

			// Check if the model exists in the available models
			if availableModel, ok := availableModelMap[baseName]; ok {
				// Parse update times
				availableUpdateTime := parseUpdateTime(availableModel.Updated)
				installedUpdateTime := installedModel.ModifiedAt

				// If the available model is newer than the installed model, it's outdated
				if availableUpdateTime.After(installedUpdateTime) {
					outdatedModels = append(outdatedModels, OutdatedModel{
						InstalledModel: installedModel,
						AvailableModel: availableModel,
					})
				}
			}
		}

		// If no outdated models are found, print a message and return
		if len(outdatedModels) == 0 {
			if filterName != "" {
				output.Default.InfoPrintln(fmt.Sprintf("No outdated models found matching '%s'.", filterName))
			} else {
				output.Default.InfoPrintln("All installed models are up to date.")
			}
			return nil
		}

		// Handle different output formats
		switch strings.ToLower(outputFormat) {
		case "json":
			return outputOutdatedJSON(outdatedModels)
		case "wide":
			return outputOutdatedWide(outdatedModels)
		default:
			return outputOutdatedTable(outdatedModels, showDetails)
		}
	},
}

// OutdatedModel represents a model that is outdated
type OutdatedModel struct {
	InstalledModel api.ListModelResponse
	AvailableModel available.Model
}

// outputOutdatedTable formats and displays the outdated models in a table format
func outputOutdatedTable(models []OutdatedModel, showDetails bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if showDetails {
		fmt.Fprintln(w, output.MakeHeader("NAME\tINSTALLED DATE\tAVAILABLE DATE\tSIZE\tPARAMETERS"))
	} else {
		fmt.Fprintln(w, output.MakeHeader("NAME\tINSTALLED DATE\tAVAILABLE DATE"))
	}

	for _, model := range models {
		if showDetails {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				output.Highlight(model.InstalledModel.Name),
				output.Warning(model.InstalledModel.ModifiedAt.Format("2006-01-02")),
				output.Success(parseUpdateTime(model.AvailableModel.Updated).Format("2006-01-02")),
				output.Info(model.AvailableModel.Size),
				output.Info(model.InstalledModel.Details.ParameterSize),
			)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\n",
				output.Highlight(model.InstalledModel.Name),
				output.Warning(model.InstalledModel.ModifiedAt.Format("2006-01-02")),
				output.Success(parseUpdateTime(model.AvailableModel.Updated).Format("2006-01-02")),
			)
		}
	}

	return w.Flush()
}

// outputOutdatedWide formats and displays the outdated models in a wide table format
func outputOutdatedWide(models []OutdatedModel) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, output.MakeHeader("NAME\tINSTALLED DATE\tAVAILABLE DATE\tSIZE\tPULLS\tTAGS\tPARAMETERS\tQUANTIZATION\tDESCRIPTION"))

	for _, model := range models {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			output.Highlight(model.InstalledModel.Name),
			output.Warning(model.InstalledModel.ModifiedAt.Format("2006-01-02")),
			output.Success(parseUpdateTime(model.AvailableModel.Updated).Format("2006-01-02")),
			output.Info(model.AvailableModel.Size),
			output.Info(model.AvailableModel.Pulls),
			output.Info(model.AvailableModel.Tags),
			output.Info(model.InstalledModel.Details.ParameterSize),
			output.Info(model.InstalledModel.Details.QuantizationLevel),
			model.AvailableModel.Description,
		)
	}

	return w.Flush()
}

// outputOutdatedJSON outputs the outdated models in JSON format
func outputOutdatedJSON(models []OutdatedModel) error {
	jsonData, err := json.MarshalIndent(models, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal models to JSON: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
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

	// If all parsing attempts fail, return current time as fallback
	// This is not ideal but prevents errors in the comparison logic
	return time.Now()
}

func init() {
	rootCmd.AddCommand(outdatedCmd)

	// Add flags for the outdated command
	outdatedCmd.Flags().StringP("output", "o", "table", "Output format (table, wide, json)")
	outdatedCmd.Flags().BoolP("details", "d", false, "Show detailed information about models")
	outdatedCmd.Flags().StringVarP(&filterName, "filter", "f", "", "Filter models by name")
	outdatedCmd.Flags().IntVarP(&timeout, "timeout", "t", 30, "Timeout in seconds for the HTTP request")
}
