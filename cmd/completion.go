package cmd

import (
	"context"
	"strings"

	"github.com/masgari/ollama-cli/pkg/client"
	"github.com/masgari/ollama-cli/pkg/config"
	"github.com/spf13/cobra"
)

// completeModelNames provides completion for model names from the Ollama server
// This function can be used by any command that requires a model name argument
func completeModelNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Get configuration context from command flags
	configName := ""
	if cmd.Flags().Changed("config-name") {
		configName, _ = cmd.Flags().GetString("config-name")
	} else if cmd.Flags().Changed("c") {
		configName, _ = cmd.Flags().GetString("c")
	}

	// Load the appropriate configuration
	cfg, err := config.LoadConfig(configName)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Override config with command line flags if provided
	if cmd.Flags().Changed("host") {
		host, _ := cmd.Flags().GetString("host")
		cfg.Host = host
	}
	if cmd.Flags().Changed("port") {
		port, _ := cmd.Flags().GetInt("port")
		cfg.Port = port
	}
	if cmd.Flags().Changed("tls") {
		tls, _ := cmd.Flags().GetBool("tls")
		cfg.Tls = tls
	}

	// Create Ollama client with the correct configuration
	ollamaClient, err := client.NewClientWithConfig(cfg)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Fetch available models
	models, err := ollamaClient.ListModels(context.Background())
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Extract model names and filter based on what user has typed
	var modelNames []string
	for _, model := range models.Models {
		if strings.HasPrefix(strings.ToLower(model.Name), strings.ToLower(toComplete)) {
			modelNames = append(modelNames, model.Name)
		}
	}

	return modelNames, cobra.ShellCompDirectiveNoFileComp
}
