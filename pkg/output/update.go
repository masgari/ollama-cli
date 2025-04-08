package output

import (
	"fmt"

	"github.com/fatih/color"
)

// ShowUpdateNotification displays a notification about available updates
func ShowUpdateNotification(currentVersion, latestVersion string) {
	// Use a different color to make it stand out
	fmt.Printf("\n%s A new version of Ollama CLI is available: %s → %s\n",
		Header("UPDATE:"),
		currentVersion,
		Success(latestVersion))

	// Create an underlined URL
	url := "https://github.com/masgari/ollama-cli/releases"
	underlinedURL := color.New(color.Underline).Sprint(url)
	fmt.Printf("Visit %s to download the latest version.\n\n", underlinedURL)
}
