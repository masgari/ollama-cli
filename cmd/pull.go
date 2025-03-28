package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/masgari/ollama-cli/pkg/output"
	"github.com/spf13/cobra"
)

// colorizeDuration applies color to a duration based on its magnitude
func colorizeDuration(duration time.Duration) string {
	// Convert duration to milliseconds for easier comparison
	durationMs := float64(duration.Milliseconds())

	// Apply colors based on duration magnitude
	if durationMs < 300000.0 { // Under 5 minutes
		// Fast (under 5m) - green
		return output.Success(duration.Round(time.Second).String())
	} else if durationMs < 900000.0 { // Under 15 minutes
		// Medium (5m - 15m) - blue
		return output.Info(duration.Round(time.Second).String())
	} else if durationMs < 1800000.0 { // Under 30 minutes
		// Slow (15m - 30m) - yellow
		return output.Warning(duration.Round(time.Second).String())
	} else {
		// Very slow (over 30m) - red
		return output.Error(duration.Round(time.Second).String())
	}
}

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull [model]",
	Short: "Pull a model from the Ollama server",
	Long:  `Pull a model and its data from the remote Ollama server.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		modelName := args[0]

		ollamaClient, err := createOllamaClient()
		if err != nil {
			return err
		}

		output.Default.InfoPrintf("Pulling model '%s'...\n", output.Highlight(modelName))
		start := time.Now()
		if err := ollamaClient.PullModel(context.Background(), modelName); err != nil {
			return fmt.Errorf("failed to pull model: %w", err)
		}
		duration := time.Since(start)

		output.Default.SuccessPrintf("\nModel '%s' pulled successfully in %s.\n", output.Highlight(modelName), colorizeDuration(duration))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
