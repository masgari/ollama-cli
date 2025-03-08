package available

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/masgari/ollama-cli/pkg/output"
)

// OutputTable formats and displays the models in a table format
func OutputTable(models []Model, showDetails bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	if showDetails {
		fmt.Fprintln(w, output.MakeHeader("NAME\tSIZE\tUPDATED\tDESCRIPTION"))
	} else {
		fmt.Fprintln(w, output.MakeHeader("NAME\tSIZE\tUPDATED"))
	}

	for _, model := range models {
		if showDetails {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				output.Highlight(model.Name),
				output.Info(formatSize(model.Size)),
				output.Info(formatUpdated(model.Updated)),
				getOrDefault(model.Description, ""),
			)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\n",
				output.Highlight(model.Name),
				output.Info(formatSize(model.Size)),
				output.Info(formatUpdated(model.Updated)),
			)
		}
	}

	return w.Flush()
}

// OutputWide formats and displays the models in a wide table format
func OutputWide(models []Model) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, output.MakeHeader("NAME\tSIZE\tPULLS\tTAGS\tUPDATED\tDESCRIPTION"))

	for _, model := range models {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			output.Highlight(model.Name),
			output.Info(formatSize(model.Size)),
			getOrDefault(model.Pulls, ""),
			getOrDefault(model.Tags, ""),
			getOrDefault(model.Updated, ""),
			getOrDefault(model.Description, ""),
		)
	}

	return w.Flush()
}

// OutputJSON outputs the models in JSON format
func OutputJSON(models []Model) error {
	jsonData, err := json.MarshalIndent(models, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal models to JSON: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

// getOrDefault returns the value if not empty, otherwise returns the default value
func getOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// formatSize simplifies the size display by showing a summary
func formatSize(size string) string {
	if size == "" {
		return ""
	}

	// Split sizes and remove duplicates
	sizes := strings.Split(size, ", ")
	seen := make(map[string]bool)
	unique := []string{}

	for _, s := range sizes {
		s = strings.TrimSpace(s)
		if s != "" && !seen[s] {
			seen[s] = true
			unique = append(unique, s)
		}
	}

	// If there are too many sizes, show a summary
	if len(unique) > 8 {
		// Try to find the smallest and largest sizes
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)(m|b)`)
		var min, max float64 = -1, -1

		for _, s := range unique {
			matches := re.FindStringSubmatch(strings.ToLower(s))
			if len(matches) == 3 {
				val := parseFloat(matches[1])
				if matches[2] == "m" {
					val = val / 1000 // convert millions to billions for comparison
				}
				if min == -1 || val < min {
					min = val
				}
				if max == -1 || val > max {
					max = val
				}
			}
		}

		if min != -1 && max != -1 {
			// Format the range, always showing in billions
			if min < 1 {
				minStr := formatFloat(min * 1000)
				maxStr := formatFloat(max)
				return fmt.Sprintf("%sm - %sb", minStr, maxStr)
			} else {
				minStr := formatFloat(min)
				maxStr := formatFloat(max)
				return fmt.Sprintf("%sb - %sb", minStr, maxStr)
			}
		}
	}

	// If 3 or fewer sizes, or if we couldn't parse the range, show all unique sizes
	return strings.Join(unique, ",")
}

// parseFloat parses a float string
func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// formatFloat formats a float without trailing zeros
func formatFloat(f float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.1f", f), "0"), ".")
}

// formatUpdated formats the update time in a readable format
func formatUpdated(updated string) string {
	if updated == "" {
		return "-"
	}

	// Parse the time
	t := parseUpdateTime(updated)
	if t.IsZero() {
		return updated // Return original string if we can't parse it
	}

	// Calculate duration since update
	duration := time.Since(t)

	// Format duration in a human-readable way
	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case duration < 7*24*time.Hour:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case duration < 30*24*time.Hour:
		weeks := int(duration.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case duration < 365*24*time.Hour:
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(duration.Hours() / 24 / 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}
