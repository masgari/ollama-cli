package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/masgari/ollama-cli/pkg/config"
	"github.com/masgari/ollama-cli/pkg/output"
	"github.com/spf13/cobra"
)

var (
	configHost string
	configPort int
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure the Ollama CLI",
	Long:  `Configure the Ollama CLI to connect to a remote Ollama server.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no flags are provided, show the current configuration
		if configHost == "" && configPort == 0 {
			output.Default.HeaderPrintln("Current configuration:")
			if configName != "" {
				fmt.Printf("  Config: %s\n", output.Highlight(configName))
			}
			fmt.Printf("  Host: %s\n", output.Highlight(cfg.Host))
			fmt.Printf("  Port: %s\n", output.Highlight(strconv.Itoa(cfg.Port)))
			fmt.Printf("  URL:  %s\n", output.Highlight(cfg.GetServerURL()))
			return
		}

		// Update the configuration
		if configHost != "" {
			cfg.Host = configHost
		}
		if configPort != 0 {
			cfg.Port = configPort
		}

		// Save the configuration
		if err := config.SaveConfig(cfg, configName); err != nil {
			output.Default.ErrorPrintf("Error saving configuration: %v\n", err)
			return
		}

		output.Default.SuccessPrintln("Configuration updated successfully:")
		if configName != "" {
			fmt.Printf("  Config: %s\n", output.Highlight(configName))
		}
		fmt.Printf("  Host: %s\n", output.Highlight(cfg.Host))
		fmt.Printf("  Port: %s\n", output.Highlight(strconv.Itoa(cfg.Port)))
		fmt.Printf("  URL:  %s\n", output.Highlight(cfg.GetServerURL()))
	},
}

// configSetCmd represents the config set command
var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long:  `Set a configuration value for the Ollama CLI.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		switch key {
		case "host":
			cfg.Host = value
		case "port":
			port, err := strconv.Atoi(value)
			if err != nil {
				output.Default.ErrorPrintln("Error: port must be a number")
				return
			}
			cfg.Port = port
		default:
			output.Default.ErrorPrintf("Error: unknown configuration key: %s\n", key)
			return
		}

		// Save the configuration
		if err := config.SaveConfig(cfg, configName); err != nil {
			output.Default.ErrorPrintf("Error saving configuration: %v\n", err)
			return
		}

		output.Default.SuccessPrintln("Configuration updated successfully:")
		fmt.Printf("  %s: %s\n", output.MakeHeader(key), output.Highlight(value))
	},
}

// configGetCmd represents the config get command
var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Long:  `Get a configuration value from the Ollama CLI.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		switch key {
		case "host":
			fmt.Println(output.Highlight(cfg.Host))
		case "port":
			fmt.Println(output.Highlight(strconv.Itoa(cfg.Port)))
		case "url":
			fmt.Println(output.Highlight(cfg.GetServerURL()))
		default:
			output.Default.ErrorPrintf("Error: unknown configuration key: %s\n", key)
		}
	},
}

// configListCmd represents the config list command
var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available configurations",
	Long:  `List all available configuration files in the Ollama CLI config directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the config directory
		configDir := config.GetConfigDir()

		// Check if the directory exists
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			output.Default.ErrorPrintln("Config directory does not exist")
			return
		}

		// List all YAML files in the directory
		files, err := os.ReadDir(configDir)
		if err != nil {
			output.Default.ErrorPrintf("Error reading config directory: %v\n", err)
			return
		}

		output.Default.HeaderPrintln("Available configurations:")

		found := false
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
				configName := strings.TrimSuffix(file.Name(), ".yaml")
				fmt.Printf("  %s\n", output.Highlight(configName))
				found = true
			}
		}

		if !found {
			fmt.Println("  No configuration files found")
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)

	// Add flags for the config command
	configCmd.Flags().StringVar(&configHost, "host", "", "Ollama server host")
	configCmd.Flags().IntVar(&configPort, "port", 0, "Ollama server port")
}
