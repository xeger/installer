package updatecheck

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/crossnokaye/cli/internal/github"
	"github.com/crossnokaye/cli/internal/version"
)

// UpdateState represents the state of the update checker
type UpdateState struct {
	CheckedAt string `json:"checked_at"`
}

const (
	// UpdateCheckInterval is the interval at which to check for updates
	UpdateCheckInterval = 7 * 24 * time.Hour
)

// getStateDir returns the XDG_STATE_DIR or appropriate fallback
func getStateDir() string {
	stateDir := os.Getenv("XDG_STATE_HOME")
	if stateDir == "" {
		homeDir, _ := os.UserHomeDir()
		stateDir = filepath.Join(homeDir, ".local", "state")
	}
	return filepath.Join(stateDir, "crossnokaye", "cli")
}

// getStateFile returns the path to the state file
func getStateFile() string {
	return filepath.Join(getStateDir(), "updatecheck.json")
}

// loadUpdateState loads the update state from the state file
func loadUpdateState() (*UpdateState, error) {
	stateFile := getStateFile()

	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &UpdateState{}, nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state UpdateState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return &state, nil
}

// saveUpdateState saves the update state to the state file
func saveUpdateState(state *UpdateState) error {
	stateDir := getStateDir()
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	stateFile := getStateFile()
	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// shouldCheckForUpdate returns true if it's been more than a week since the last check
func shouldCheckForUpdate(state *UpdateState) bool {
	if state.CheckedAt == "" {
		return true
	}

	lastCheck, err := time.Parse(time.RFC3339, state.CheckedAt)
	if err != nil {
		return true // If we can't parse the time, assume we should check
	}

	return time.Since(lastCheck) > UpdateCheckInterval
}

// compareVersions compares two version strings and returns true if newVersion is newer than currentVersion
func compareVersions(currentVersion, newVersion string) bool {
	// Simple string comparison for now - this handles semantic versioning reasonably well
	// since "v1.10.0" > "v1.9.0" lexicographically
	if currentVersion == "" {
		return true
	}

	// If both versions start with 'v', compare them directly
	if strings.HasPrefix(currentVersion, "v") && strings.HasPrefix(newVersion, "v") {
		return newVersion > currentVersion
	}

	// If neither starts with 'v', compare directly
	if !strings.HasPrefix(currentVersion, "v") && !strings.HasPrefix(newVersion, "v") {
		return newVersion > currentVersion
	}

	// Handle mixed cases by normalizing
	current := strings.TrimPrefix(currentVersion, "v")
	new := strings.TrimPrefix(newVersion, "v")
	return new > current
}

// CheckForUpdates checks for updates and displays a message if a newer version is available
func CheckForUpdates() {
	state, err := loadUpdateState()
	if err != nil {
		// Silently fail if we can't load state
		return
	}

	if !shouldCheckForUpdate(state) {
		return
	}

	// Update the last check time
	state.CheckedAt = time.Now().Format(time.RFC3339)
	if err := saveUpdateState(state); err != nil {
		// Silently fail if we can't save state
		return
	}

	// Skip update check if ReleaseChannel is empty
	if version.ReleaseChannel == "" {
		return
	}

	// Check for updates
	release, err := github.GetLatestRelease(version.ReleaseChannel)
	if err != nil {
		// Silently fail if we can't get release info
		return
	}

	currentVersion := version.ReleaseTag
	if compareVersions(currentVersion, release.TagName) {
		fmt.Printf("\n🎉 Update to CLI %s at https://github.com/%s/releases/latest\n", release.TagName, version.ReleaseChannel)
	}
}
