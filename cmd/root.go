package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/masgari/ollama-cli/pkg/config"
	"github.com/masgari/ollama-cli/pkg/output"
	"github.com/masgari/ollama-cli/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	configName string
	noColor    bool
	verbose    bool
	noUpdates  bool
)

// GetConfig returns the current configuration
func GetConfig() *config.Config { return config.Current }

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ollama-cli",
	Short: "A CLI tool for interacting with a remote Ollama server",
	Long: `ollama-cli is a command-line interface for interacting with a remote Ollama server.
It allows you to manage models, run inferences, and more.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Handle color flag
		if noColor {
			output.DisableColors()
		}

		// Set the global configuration name
		config.CurrentConfigName = configName

		var err error
		loadedCfg, err := config.LoadConfig(configName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Override config with command line flags if provided
		if cmd.Flags().Changed("host") {
			host, _ := cmd.Flags().GetString("host")
			loadedCfg.Host = host
		}
		if cmd.Flags().Changed("port") {
			port, _ := cmd.Flags().GetInt("port")
			loadedCfg.Port = port
		}
		if cmd.Flags().Changed("tls") {
			tls, _ := cmd.Flags().GetBool("tls")
			loadedCfg.Tls = tls
		}

		// Expose the final, effective configuration to the client factory.
		config.Current = loadedCfg

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Check for updates if enabled in config and not disabled by flag
		if config.Current != nil && config.Current.CheckUpdates && !noUpdates {
			hasUpdate, current, latest, err := version.CheckForUpdates(Version)
			if err == nil && hasUpdate {
				output.ShowUpdateNotification(current, latest)
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ollama-cli/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&configName, "config-name", "c", "", "config name to use (e.g. 'pc' for $HOME/.ollama-cli/pc.yaml)")
	rootCmd.PersistentFlags().StringP("host", "H", "", "Ollama server host (default is localhost)")
	rootCmd.PersistentFlags().Int("port", 0, "Ollama server port (default is 11434)")
	rootCmd.PersistentFlags().Bool("tls", false, "Use TLS for Ollama server connection")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&noUpdates, "no-updates", false, "Disable update checks")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".ollama-cli" (without extension).
		viper.AddConfigPath(home + "/.ollama-cli")
		viper.SetConfigType("yaml")

		// If configName is provided, use that as the config name
		if configName != "" {
			viper.SetConfigName(configName)
		} else {
			viper.SetConfigName("config")
		}
	}

	// Only read environment variables prefixed with OLLAMA_CLI_
	// e.g., OLLAMA_CLI_HOST, OLLAMA_CLI_PORT, OLLAMA_CLI_TLS, etc.
	viper.SetEnvPrefix("OLLAMA_CLI")
	// Normalize keys to env var form (dots/hyphens to underscores)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}
}
