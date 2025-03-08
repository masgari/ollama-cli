package cmd

import (
	"fmt"
	"os"

	"github.com/masgari/ollama-cli/pkg/config"
	"github.com/masgari/ollama-cli/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	cfg     *config.Config
	noColor bool
	verbose bool
)

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

		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
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

		return nil
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
	rootCmd.PersistentFlags().StringP("host", "H", "", "Ollama server host (default is localhost)")
	rootCmd.PersistentFlags().IntP("port", "p", 0, "Ollama server port (default is 11434)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
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
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}
}
