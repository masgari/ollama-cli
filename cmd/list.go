package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/masgari/ollama-cli/pkg/client"
	"github.com/masgari/ollama-cli/pkg/output"
	"github.com/ollama/ollama/api"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	showDetails  bool
	timeNow      = time.Now // For testing
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List models available on the Ollama server",
	Long:    `List all models that are available on the remote Ollama server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get a client using the factory approach
		ollamaClient := client.NewClient()

		models, err := ollamaClient.ListModels(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list models: %w", err)
		}

		// If no models are available, print a message and return
		if len(models.Models) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No models found on the Ollama server.")
			return nil
		}

		// Handle different output formats
		switch strings.ToLower(outputFormat) {
		case "json":
			return outputJSON(cmd.OutOrStdout(), models)
		case "wide":
			return outputWide(cmd.OutOrStdout(), models)
		case "table":
			return outputTable(cmd.OutOrStdout(), models, showDetails)
		default:
			return fmt.Errorf("invalid output format: %s", outputFormat)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Add flags for the list command
	listCmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, wide, json)")
	listCmd.Flags().BoolVarP(&showDetails, "details", "d", false, "Show detailed information about models")
}

// outputTable formats and displays the models in a table format
func outputTable(out io.Writer, models *api.ListResponse, showDetails bool) error {
	w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)

	if showDetails {
		fmt.Fprintln(w, output.MakeHeader("NAME\tSIZE\tMODIFIED\tQUANTIZATION\tFAMILY\tPARAMETERS"))
	} else {
		fmt.Fprintln(w, output.MakeHeader("NAME\tSIZE\tMODIFIED"))
	}

	for _, model := range models.Models {
		if showDetails {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				output.Highlight(model.Name),
				formatSize(model.Size),
				formatTime(model.ModifiedAt),
				getOrDefault(model.Details.QuantizationLevel, "N/A"),
				getOrDefault(model.Details.Family, "N/A"),
				getOrDefault(model.Details.ParameterSize, "N/A"),
			)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\n",
				output.Highlight(model.Name),
				formatSize(model.Size),
				formatTime(model.ModifiedAt),
			)
		}
	}

	return w.Flush()
}

// outputWide formats and displays the models in a wide table format with all details
func outputWide(out io.Writer, models *api.ListResponse) error {
	w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, output.MakeHeader("NAME\tSIZE\tMODIFIED\tQUANTIZATION\tFAMILY\tPARAMETERS\tDIGEST"))

	for _, model := range models.Models {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			output.Highlight(model.Name),
			formatSize(model.Size),
			formatTime(model.ModifiedAt),
			getOrDefault(model.Details.QuantizationLevel, "N/A"),
			getOrDefault(model.Details.Family, "N/A"),
			getOrDefault(model.Details.ParameterSize, "N/A"),
			getOrDefault(model.Digest, "N/A"),
		)
	}

	return w.Flush()
}

// outputJSON outputs the models in JSON format
func outputJSON(out io.Writer, models *api.ListResponse) error {
	jsonData, err := json.MarshalIndent(models, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal models to JSON: %w", err)
	}

	fmt.Fprintln(out, string(jsonData))
	return nil
}

// formatSize formats the size in bytes to a human-readable format
func formatSize(sizeInBytes int64) string {
	const (
		_          = iota
		KB float64 = 1 << (10 * iota)
		MB
		GB
	)

	size := float64(sizeInBytes)

	switch {
	case size >= GB:
		return output.Info(fmt.Sprintf("%.1f GB", size/GB))
	case size >= MB:
		return output.Info(fmt.Sprintf("%.1f MB", size/MB))
	case size >= KB:
		return output.Info(fmt.Sprintf("%.1f KB", size/KB))
	default:
		return output.Info(fmt.Sprintf("%d B", sizeInBytes))
	}
}

// formatTime formats the time to a human-readable format
func formatTime(t time.Time) string {
	now := timeNow()
	diff := now.Sub(t)

	switch {
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		return output.Warning(fmt.Sprintf("%d minutes ago", minutes))
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return output.Warning(fmt.Sprintf("%d hours ago", hours))
	case diff < 30*24*time.Hour:
		days := int(diff.Hours() / 24)
		return output.Warning(fmt.Sprintf("%d days ago", days))
	default:
		months := int(diff.Hours() / 24 / 30)
		return output.Warning(fmt.Sprintf("%d months ago", months))
	}
}

// getOrDefault returns the value or a default if the value is empty
func getOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
