package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List models available on the Ollama server",
	Long:    `List all models that are available on the remote Ollama server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ollamaClient, err := client.New(cfg)
		if err != nil {
			return fmt.Errorf("failed to create Ollama client: %w", err)
		}

		models, err := ollamaClient.ListModels(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list models: %w", err)
		}

		// If no models are available, print a message and return
		if len(models.Models) == 0 {
			output.Default.InfoPrintln("No models found on the Ollama server.")
			return nil
		}

		// Handle different output formats
		switch strings.ToLower(outputFormat) {
		case "json":
			return outputJSON(models)
		case "wide":
			return outputWide(models)
		default:
			return outputTable(models, showDetails)
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
func outputTable(models *api.ListResponse, showDetails bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

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
func outputWide(models *api.ListResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
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
func outputJSON(models *api.ListResponse) error {
	jsonData, err := json.MarshalIndent(models, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal models to JSON: %w", err)
	}

	fmt.Println(string(jsonData))
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
		return output.Info(fmt.Sprintf("%.2f GB", size/GB))
	case size >= MB:
		return output.Info(fmt.Sprintf("%.2f MB", size/MB))
	case size >= KB:
		return output.Info(fmt.Sprintf("%.2f KB", size/KB))
	default:
		return output.Info(fmt.Sprintf("%d B", sizeInBytes))
	}
}

// formatTime formats the time to a human-readable format
func formatTime(t time.Time) string {
	return output.Warning(t.Format("2006-01-02 15:04:05"))
}

// getOrDefault returns the value or a default if the value is empty
func getOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
