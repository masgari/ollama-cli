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
	configHost         string
	configPort         int
	configTls          bool
	configCheckUpdates bool
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure the Ollama CLI",
	Long: `Configure the Ollama CLI.
	
You can view or update the configuration for the Ollama CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If flags are provided, update the configuration
		if cmd.Flags().Changed("host") || cmd.Flags().Changed("port") || cmd.Flags().Changed("check-updates") {
			// Update the configuration
			if cmd.Flags().Changed("host") {
				cfg.Host = configHost
			}
			if cmd.Flags().Changed("port") {
				cfg.Port = configPort
			}
			if cmd.Flags().Changed("tls") {
				cfg.Tls = configTls
			}
			if cmd.Flags().Changed("check-updates") {
				cfg.CheckUpdates = configCheckUpdates
			}

			// Save the configuration
			if err := config.SaveConfig(cfg, configName); err != nil {
				output.Default.ErrorPrintf("Error saving configuration: %v\n", err)
				return
			}

			output.Default.SuccessPrintln("Configuration updated successfully:")
		}

		// Display the current configuration
		output.Default.HeaderPrintln("Current configuration:")
		fmt.Printf("  %s: %s\n", output.MakeHeader("Host"), output.Highlight(cfg.Host))
		fmt.Printf("  %s: %s\n", output.MakeHeader("Port"), output.Highlight(strconv.Itoa(cfg.Port)))
		fmt.Printf("  %s: %s\n", output.MakeHeader("Tls"), output.Highlight(strconv.FormatBool(cfg.Tls)))
		fmt.Printf("  %s: %s\n", output.MakeHeader("URL"), output.Highlight(cfg.GetServerURL()))
		fmt.Printf("  %s: %s\n", output.MakeHeader("Chat Enabled"), output.Highlight(strconv.FormatBool(cfg.ChatEnabled)))
		fmt.Printf("  %s: %s\n", output.MakeHeader("Check Updates"), output.Highlight(strconv.FormatBool(cfg.CheckUpdates)))
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
		case "tls":
			tls, err := strconv.ParseBool(value)
			if err != nil {
				output.Default.ErrorPrintln("Error: tls must be a boolean (true/false)")
				return
			}
			cfg.Tls = tls
		case "check-updates":
			checkUpdates, err := strconv.ParseBool(value)
			if err != nil {
				output.Default.ErrorPrintln("Error: check-updates must be a boolean (true/false)")
				return
			}
			cfg.CheckUpdates = checkUpdates
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
		case "tls":
			fmt.Println(output.Highlight(strconv.FormatBool(cfg.Tls)))
		case "url":
			fmt.Println(output.Highlight(cfg.GetServerURL()))
		case "chat_enabled":
			fmt.Println(output.Highlight(strconv.FormatBool(cfg.ChatEnabled)))
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

// configEnableChatCmd represents the config enable-chat command
var configEnableChatCmd = &cobra.Command{
	Use:   "enable-chat",
	Short: "Enable the chat command",
	Long:  `Enable the chat command in the configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg.ChatEnabled = true
		if err := config.SaveConfig(cfg, configName); err != nil {
			output.Default.ErrorPrintf("Error saving configuration: %v\n", err)
			return
		}
		output.Default.SuccessPrintf("Chat command has been enabled in your configuration.\n")
	},
}

// configDisableChatCmd represents the config disable-chat command
var configDisableChatCmd = &cobra.Command{
	Use:   "disable-chat",
	Short: "Disable the chat command",
	Long:  `Disable the chat command in the configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg.ChatEnabled = false
		if err := config.SaveConfig(cfg, configName); err != nil {
			output.Default.ErrorPrintf("Error saving configuration: %v\n", err)
			return
		}
		output.Default.SuccessPrintf("Chat command has been disabled in your configuration.\n")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configEnableChatCmd)
	configCmd.AddCommand(configDisableChatCmd)

	// Add flags for the config command
	configCmd.Flags().StringVar(&configHost, "host", "", "Ollama server host")
	configCmd.Flags().IntVar(&configPort, "port", 0, "Ollama server port")
	configCmd.Flags().BoolVar(&configTls, "tls", false, "Use TLS for Ollama server connection")
	configCmd.Flags().BoolVar(&configCheckUpdates, "check-updates", true, "Check for updates")
}
