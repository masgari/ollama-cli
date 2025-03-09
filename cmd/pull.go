package cmd

import (
	"context"
	"fmt"

	"github.com/masgari/ollama-cli/pkg/output"
	"github.com/spf13/cobra"
)

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
		if err := ollamaClient.PullModel(context.Background(), modelName); err != nil {
			return fmt.Errorf("failed to pull model: %w", err)
		}

		output.Default.SuccessPrintf("Model '%s' pulled successfully.\n", output.Highlight(modelName))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
