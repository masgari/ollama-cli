package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-version"
)

const (
	// GitHubAPIURL is the base URL for GitHub API
	GitHubAPIURL = "https://api.github.com/repos/masgari/ollama-cli/releases/latest"
	// CacheFileName is the name of the cache file
	CacheFileName = "version_cache.json"
	// CacheExpiration is the duration for which the cache is valid
	CacheExpiration = 24 * time.Hour
)

// VersionInfo contains information about the latest version
type VersionInfo struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// CacheEntry represents a cached version check result
type CacheEntry struct {
	LatestVersion string    `json:"latest_version"`
	CheckedAt     time.Time `json:"checked_at"`
}

// CheckForUpdates checks if a newer version is available
func CheckForUpdates(currentVersion string) (bool, string, string, error) {
	// Skip check if current version is "dev"
	if currentVersion == "dev" {
		return false, "", "", nil
	}

	// Check cache first
	cachedVersion, err := getCachedVersion()
	if err == nil && cachedVersion != "" {
		// Compare versions
		hasUpdate, err := compareVersions(currentVersion, cachedVersion)
		if err == nil {
			return hasUpdate, currentVersion, cachedVersion, nil
		}
	}

	// Fetch latest version from GitHub
	latestVersion, err := fetchLatestVersion()
	if err != nil {
		return false, currentVersion, "", err
	}

	// Cache the result
	cacheVersion(latestVersion)

	// Compare versions
	hasUpdate, err := compareVersions(currentVersion, latestVersion)
	if err != nil {
		return false, currentVersion, latestVersion, err
	}

	return hasUpdate, currentVersion, latestVersion, nil
}

// compareVersions compares two version strings
func compareVersions(current, latest string) (bool, error) {
	currentVer, err := version.NewVersion(current)
	if err != nil {
		return false, fmt.Errorf("invalid current version: %w", err)
	}

	latestVer, err := version.NewVersion(latest)
	if err != nil {
		return false, fmt.Errorf("invalid latest version: %w", err)
	}

	return currentVer.LessThan(latestVer), nil
}

// fetchLatestVersion fetches the latest version from GitHub
func fetchLatestVersion() (string, error) {
	resp, err := http.Get(GitHubAPIURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch latest version: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var release VersionInfo
	if err := json.Unmarshal(body, &release); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return release.TagName, nil
}

// getCachedVersion retrieves the cached version information
func getCachedVersion() (string, error) {
	cacheFile := getCacheFilePath()
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return "", fmt.Errorf("cache file does not exist")
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return "", fmt.Errorf("failed to read cache file: %w", err)
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return "", fmt.Errorf("failed to parse cache file: %w", err)
	}

	// Check if cache is expired
	if time.Since(entry.CheckedAt) > CacheExpiration {
		return "", fmt.Errorf("cache expired")
	}

	return entry.LatestVersion, nil
}

// cacheVersion caches the version information
func cacheVersion(version string) error {
	entry := CacheEntry{
		LatestVersion: version,
		CheckedAt:     time.Now(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	cacheFile := getCacheFilePath()
	cacheDir := filepath.Dir(cacheFile)

	// Create cache directory if it doesn't exist
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return fmt.Errorf("failed to create cache directory: %w", err)
		}
	}

	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// getCacheFilePath returns the path to the cache file
func getCacheFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory
		return CacheFileName
	}

	return filepath.Join(homeDir, ".ollama-cli", CacheFileName)
}
