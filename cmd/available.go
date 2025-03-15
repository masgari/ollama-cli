package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/masgari/ollama-cli/pkg/available"
	"github.com/masgari/ollama-cli/pkg/output"
	"github.com/spf13/cobra"
)

var (
	filterName string
	timeout    int
	limit      int
	maxSize    float64
)

// availableCmd represents the available command
var availableCmd = &cobra.Command{
	Use:     "available",
	Aliases: []string{"avail"},
	Short:   "List models available on ollama.com",
	Long:    `List all models that are available on ollama.com/search.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the output format from flags
		outputFormat, _ := cmd.Flags().GetString("output")
		showDetails, _ := cmd.Flags().GetBool("details")

		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()

		// Create HTTP client with timeout
		client := &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		}

		// Create ModelFetcher with the client
		fetcher := available.NewModelFetcher(client, "https://ollama.com/search")

		// Fetch available models using the fetcher
		models, err := fetcher.FetchModels(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch available models: %w", err)
		}

		// Filter models if filter is provided
		models = available.FilterByName(models, filterName)

		// Filter models by size if maxSize is provided
		models = available.FilterBySize(models, maxSize)

		// Create a custom output writer that writes to the command's output buffer
		out := output.NewColorWriter(cmd.OutOrStdout())

		// If no models are available, print a message and return
		if len(models) == 0 {
			if filterName != "" && maxSize > 0 {
				out.InfoPrintln(fmt.Sprintf("No models found matching '%s' with size <= %gb on ollama.com.", filterName, maxSize))
			} else if filterName != "" {
				out.InfoPrintln(fmt.Sprintf("No models found matching '%s' on ollama.com.", filterName))
			} else if maxSize > 0 {
				out.InfoPrintln(fmt.Sprintf("No models found with size <= %gb on ollama.com.", maxSize))
			} else {
				out.InfoPrintln("No models found on ollama.com.")
			}
			return nil
		}

		// Store the total count before applying limit
		totalCount := len(models)

		// Apply limit if specified and valid
		if limit > 0 && limit < len(models) {
			models = models[:limit]
		}

		// Handle different output formats
		var outputErr error
		switch strings.ToLower(outputFormat) {
		case "json":
			outputErr = available.OutputJSONWithWriter(cmd.OutOrStdout(), models)
		case "wide":
			outputErr = available.OutputWideWithWriter(cmd.OutOrStdout(), models)
		default:
			outputErr = available.OutputTableWithWriter(cmd.OutOrStdout(), models, showDetails)
		}

		// If we limited the output, show a message about how many models were displayed
		if limit > 0 && limit < totalCount {
			out.InfoPrintln(fmt.Sprintf("Displaying %d of %d models. Use --limit=-1 to show all.", limit, totalCount))
		}

		return outputErr
	},
}

func init() {
	rootCmd.AddCommand(availableCmd)

	// Add flags for the available command
	availableCmd.Flags().StringP("output", "o", "table", "Output format (table, wide, json)")
	availableCmd.Flags().BoolP("details", "d", false, "Show detailed information about models")
	availableCmd.Flags().StringVarP(&filterName, "filter", "f", "", "Filter models by name")
	availableCmd.Flags().IntVarP(&timeout, "timeout", "t", 30, "Timeout in seconds for the HTTP request")
	availableCmd.Flags().IntVarP(&limit, "limit", "l", 10, "Limit the number of models displayed (-1 for all)")
	availableCmd.Flags().Float64VarP(&maxSize, "size", "s", 0, "Filter models by maximum size in billions (e.g., 7 for 7B models)")
}
