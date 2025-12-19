package update

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	githubAPIURL    = "https://api.github.com/repos/DylanDevelops/tmpo/releases/latest"
	checkTimeout    = 3 * time.Second
	connectTimeout  = 2 * time.Second
)

// ReleaseInfo represents the GitHub release information
type ReleaseInfo struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
}

// UpdateInfo contains information about available updates
type UpdateInfo struct {
	CurrentVersion string
	LatestVersion  string
	UpdateURL      string
	HasUpdate      bool
}

// IsConnectedToInternet performs a quick check to see if the user has internet connectivity.
// It tries to resolve a reliable DNS name (GitHub's) with a short timeout.
func IsConnectedToInternet() bool {
	// Try to resolve GitHub's DNS with a 2-second timeout
	_, err := net.LookupHost("api.github.com")
	return err == nil
}

// GetLatestVersion fetches the latest release version from GitHub API.
// It returns the version tag (e.g., "v1.2.3") and any error encountered.
func GetLatestVersion() (string, error) {
	client := &http.Client{
		Timeout: checkTimeout,
	}

	resp, err := client.Get(githubAPIURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse release info: %w", err)
	}

	return release.TagName, nil
}

// CompareVersions compares two semantic version strings.
// Returns:
//   -1 if current < latest (update available)
//    0 if current == latest (up to date)
//    1 if current > latest (ahead of latest, e.g., dev build)
//
// Handles versions with or without "v" prefix (v1.2.3 or 1.2.3)
func CompareVersions(current, latest string) int {
	// Remove "v" prefix if present
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	// Split into parts
	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	// Compare each part
	maxLen := len(currentParts)
	if len(latestParts) > maxLen {
		maxLen = len(latestParts)
	}

	for i := 0; i < maxLen; i++ {
		var currentVal, latestVal int

		if i < len(currentParts) {
			fmt.Sscanf(currentParts[i], "%d", &currentVal)
		}
		if i < len(latestParts) {
			fmt.Sscanf(latestParts[i], "%d", &latestVal)
		}

		if currentVal < latestVal {
			return -1
		}
		if currentVal > latestVal {
			return 1
		}
	}

	return 0
}

// CheckForUpdate checks if a newer version is available.
// It first verifies internet connectivity, then fetches the latest version from GitHub.
// Returns UpdateInfo with details about the update, or an error if the check fails.
func CheckForUpdate(currentVersion string) (*UpdateInfo, error) {
	info := &UpdateInfo{
		CurrentVersion: currentVersion,
		HasUpdate:      false,
	}

	// Quick internet connectivity check
	if !IsConnectedToInternet() {
		return nil, fmt.Errorf("no internet connection")
	}

	// Fetch latest version from GitHub
	latestVersion, err := GetLatestVersion()
	if err != nil {
		return nil, err
	}

	info.LatestVersion = latestVersion
	info.UpdateURL = fmt.Sprintf("https://github.com/DylanDevelops/tmpo/releases/tag/%s", latestVersion)

	// Compare versions
	comparison := CompareVersions(currentVersion, latestVersion)
	if comparison < 0 {
		info.HasUpdate = true
	}

	return info, nil
}
