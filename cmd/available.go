package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/masgari/ollama-cli/pkg/available"
	"github.com/masgari/ollama-cli/pkg/output"
	"github.com/spf13/cobra"
)

var (
	filterName string
	timeout    int
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

		// Fetch available models
		models, err := available.FetchModels(ctx, timeout)
		if err != nil {
			return fmt.Errorf("failed to fetch available models: %w", err)
		}

		// Filter models if filter is provided
		models = available.FilterByName(models, filterName)

		// If no models are available, print a message and return
		if len(models) == 0 {
			if filterName != "" {
				output.Default.InfoPrintln(fmt.Sprintf("No models found matching '%s' on ollama.com.", filterName))
			} else {
				output.Default.InfoPrintln("No models found on ollama.com.")
			}
			return nil
		}

		// Handle different output formats
		switch strings.ToLower(outputFormat) {
		case "json":
			return available.OutputJSON(models)
		case "wide":
			return available.OutputWide(models)
		default:
			return available.OutputTable(models, showDetails)
		}
	},
}

func init() {
	rootCmd.AddCommand(availableCmd)

	// Add flags for the available command
	availableCmd.Flags().StringP("output", "o", "table", "Output format (table, wide, json)")
	availableCmd.Flags().BoolP("details", "d", false, "Show detailed information about models")
	availableCmd.Flags().StringVarP(&filterName, "filter", "f", "", "Filter models by name")
	availableCmd.Flags().IntVarP(&timeout, "timeout", "t", 30, "Timeout in seconds for the HTTP request")
}
