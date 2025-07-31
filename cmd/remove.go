package cmd

import (
	"context"
	"fmt"

	"github.com/masgari/ollama-cli/pkg/output"
	"github.com/spf13/cobra"
)

var (
	forceDelete bool
)

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:               "rm [model]",
	Aliases:           []string{"delete", "remove"},
	Short:             "Remove a model from the Ollama server",
	Long:              `Remove a model and its data from the remote Ollama server.`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeModelNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		modelName := args[0]

		ollamaClient, err := createOllamaClient()
		if err != nil {
			return err
		}

		// Check if the model exists
		models, err := ollamaClient.ListModels(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list models: %w", err)
		}

		modelExists := false
		for _, model := range models.Models {
			if model.Name == modelName {
				modelExists = true
				break
			}
		}

		if !modelExists {
			return fmt.Errorf("model '%s' not found on the server", modelName)
		}

		// Confirm deletion if not forced
		if !forceDelete {
			fmt.Fprintf(cmd.OutOrStdout(), "Are you sure you want to delete model '%s'? (y/N): ", output.Highlight(modelName))
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				output.Default.WarningPrintln("Deletion cancelled.")
				return nil
			}
		}

		if err := ollamaClient.DeleteModel(context.Background(), modelName); err != nil {
			return fmt.Errorf("failed to delete model: %w", err)
		}

		output.Default.SuccessPrintf("Model '%s' deleted successfully.\n", output.Highlight(modelName))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)

	// Add flags for the rm command
	rmCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Force deletion without confirmation")
}
