package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Release struct {
	TagName string  `json:"tagName"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// CLIRepository is the GitHub repository for the CLI itself
const CLIRepository = "xeger/ck"

// FindRepositories returns all possible repository patterns for a command name
func FindRepositories(commandName string) []string {
	return []string{
		fmt.Sprintf("crossnokaye/%s", commandName),
		fmt.Sprintf("crossnokaye/cli-%s", commandName),
		fmt.Sprintf("crossnokaye/%s-cli", commandName),
	}
}

func IsGHCLIAvailable() error {
	_, err := exec.LookPath("gh")
	if err != nil {
		return fmt.Errorf("GitHub CLI not found. Please install from https://cli.github.com/")
	}
	return nil
}

func GetLatestRelease(repo string) (*Release, error) {
	if err := IsGHCLIAvailable(); err != nil {
		return nil, err
	}

	cmd := exec.Command("gh", "release", "view", "--repo", repo, "--json", "tagName,assets")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get release info for %s: %w", repo, err)
	}

	var release Release
	if err := json.Unmarshal(output, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release data: %w", err)
	}

	return &release, nil
}

func FindAssetForPlatform(release *Release, baseName string) (*Asset, error) {
	expectedName := fmt.Sprintf("%s_%s_%s_%s.tar.gz", baseName, release.TagName, runtime.GOOS, runtime.GOARCH)

	for _, asset := range release.Assets {
		if asset.Name == expectedName {
			return &asset, nil
		}
	}

	var availableAssets []string
	for _, asset := range release.Assets {
		if strings.HasPrefix(asset.Name, baseName+"_") && strings.HasSuffix(asset.Name, ".tar.gz") {
			availableAssets = append(availableAssets, asset.Name)
		}
	}

	if len(availableAssets) > 0 {
		return nil, fmt.Errorf("no asset found for %s. Available assets: %s",
			expectedName, strings.Join(availableAssets, ", "))
	}

	return nil, fmt.Errorf("no compatible assets found for %s", baseName)
}

func DownloadAsset(asset *Asset, destPath string, repo string, tagName string) error {
	if err := IsGHCLIAvailable(); err != nil {
		return err
	}

	destDir := filepath.Dir(destPath)
	cmd := exec.Command("gh", "release", "download", tagName, "--pattern", asset.Name, "--dir", destDir, "--repo", repo)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download asset %s: %w", asset.Name, err)
	}

	return nil
}

// GetCurrentGitHubUser returns the GitHub username of the currently logged-in user
func GetCurrentGitHubUser() (string, error) {
	if err := IsGHCLIAvailable(); err != nil {
		return "", err
	}

	cmd := exec.Command("gh", "auth", "status", "--hostname", "github.com")
	output, err := cmd.Output()
	if err != nil {
		return "", nil // Not logged in or auth failed
	}

	// Parse the output to extract the username
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Logged in to github.com account") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "account" && i+1 < len(parts) {
					// Remove any parentheses or other characters
					username := strings.TrimSpace(parts[i+1])
					username = strings.Trim(username, "()")
					return username, nil
				}
			}
		}
	}

	return "", nil // Couldn't parse username
}

// FormatRepositoryNotFoundError creates a user-friendly error message when no repository is found
func FormatRepositoryNotFoundError(commandName string, repoPatterns []string) string {
	var message strings.Builder

	message.WriteString("No distribution channel found.\n")
	message.WriteString("Tried the following repositories:\n")

	for _, pattern := range repoPatterns {
		message.WriteString(fmt.Sprintf("  • %s\n", pattern))
	}

	// Get current GitHub user
	username, err := GetCurrentGitHubUser()
	if err != nil {
		message.WriteString("\nNote: GitHub CLI is not available. Please install it from https://cli.github.com/\n")
	} else if username == "" {
		message.WriteString("\nIf the command is located in a private repository, please log in with:\n")
		message.WriteString("  gh auth login\n")
	} else {
		message.WriteString(fmt.Sprintf("\nIf the command is located in a private repository, make sure that you (%s) have access to it.\n", username))
	}

	return message.String()
}
