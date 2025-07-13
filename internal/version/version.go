package version

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type ReleaseInfo struct {
	TagName    string `json:"tag_name"`
	Repository string `json:"repository"`
	Timestamp  string `json:"timestamp"`
}

type VersionDatabase map[string]ReleaseInfo

func WriteReleaseInfo(cmdDir, commandName, tagName, repository string) error {
	releaseFile := filepath.Join(cmdDir, "release.json")
	
	var db VersionDatabase
	if data, err := os.ReadFile(releaseFile); err == nil {
		json.Unmarshal(data, &db)
	} else {
		db = make(VersionDatabase)
	}

	db[commandName] = ReleaseInfo{
		TagName:    tagName,
		Repository: repository,
		Timestamp:  fmt.Sprintf("%d", time.Now().Unix()),
	}

	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal release info: %w", err)
	}

	if err := os.WriteFile(releaseFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write release info: %w", err)
	}

	return nil
}

func GetInstalledVersion(cmdDir, commandName string) (string, error) {
	releaseFile := filepath.Join(cmdDir, "release.json")
	
	data, err := os.ReadFile(releaseFile)
	if err != nil {
		return "", err
	}

	var db VersionDatabase
	if err := json.Unmarshal(data, &db); err != nil {
		return "", fmt.Errorf("failed to parse release info: %w", err)
	}

	if info, exists := db[commandName]; exists {
		return info.TagName, nil
	}

	return "", fmt.Errorf("no version info found for %s", commandName)
}

func GetAllInstalledVersions(stateHome string) (map[string]string, error) {
	cmdBaseDir := filepath.Join(stateHome, "crossnokaye", "cli", "cmd")
	versions := make(map[string]string)

	if _, err := os.Stat(cmdBaseDir); os.IsNotExist(err) {
		return versions, nil
	}

	entries, err := os.ReadDir(cmdBaseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read commands directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		commandName := entry.Name()
		cmdDir := filepath.Join(cmdBaseDir, commandName)
		
		// Only include if binary exists
		binaryPath := filepath.Join(cmdDir, commandName)
		if _, err := os.Stat(binaryPath); err != nil {
			continue
		}

		if version, err := GetInstalledVersion(cmdDir, commandName); err == nil {
			versions[commandName] = version
		}
	}

	return versions, nil
}

func GetSortedCommandsWithVersions(stateHome string) ([]string, map[string]string, error) {
	cmdBaseDir := filepath.Join(stateHome, "crossnokaye", "cli", "cmd")
	versions := make(map[string]string)
	var commands []string

	if _, err := os.Stat(cmdBaseDir); os.IsNotExist(err) {
		return commands, versions, nil
	}

	entries, err := os.ReadDir(cmdBaseDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read commands directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		commandName := entry.Name()
		cmdDir := filepath.Join(cmdBaseDir, commandName)
		
		// Only include if binary exists
		binaryPath := filepath.Join(cmdDir, commandName)
		if _, err := os.Stat(binaryPath); err != nil {
			continue
		}

		// Always add command to list
		commands = append(commands, commandName)
		
		// Add version info if available
		if version, err := GetInstalledVersion(cmdDir, commandName); err == nil {
			versions[commandName] = version
		}
	}

	sort.Strings(commands)
	return commands, versions, nil
}