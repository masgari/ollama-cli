package cmd

import (
	"fmt"

	"github.com/masgari/ollama-cli/pkg/output"
	"github.com/spf13/cobra"
)

var (
	// Version is the version of the CLI tool
	Version = "0.1.0"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the version of the CLI tool",
	Long:  `Display the version of the Ollama CLI tool.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Ollama CLI version %s\n", output.Highlight(Version))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
