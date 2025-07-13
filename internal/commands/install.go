package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/crossnokaye/cli/internal/archive"
	"github.com/crossnokaye/cli/internal/github"
	"github.com/crossnokaye/cli/internal/version"
)

func Install(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: install <command-name>")
	}

	commandName := args[0]
	repoPatterns := []string{
		fmt.Sprintf("crossnokaye/cli-%s", commandName),
		fmt.Sprintf("crossnokaye/%s", commandName),
	}

	var release *github.Release
	var repo string
	var err error

	for _, pattern := range repoPatterns {
		fmt.Printf("Trying repository %s...\n", pattern)
		release, err = github.GetLatestRelease(pattern)
		if err == nil {
			repo = pattern
			break
		}
		fmt.Printf("Repository %s not found, trying next pattern...\n", pattern)
	}

	if release == nil {
		return fmt.Errorf("failed to find repository for %s. Tried: %v", commandName, repoPatterns)
	}

	fmt.Printf("Installing %s from %s...\n", commandName, repo)

	fmt.Printf("Found release %s\n", release.TagName)

	asset, err := github.FindAssetForPlatform(release, commandName)
	if err != nil {
		return fmt.Errorf("failed to find compatible asset: %w", err)
	}

	fmt.Printf("Downloading %s...\n", asset.Name)

	stateHome := getStateHome()
	cmdDir := filepath.Join(stateHome, "crossnokaye", "cli", "cmd", commandName)
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		return fmt.Errorf("failed to create command directory: %w", err)
	}

	tempArchive := filepath.Join(cmdDir, asset.Name)
	if err := github.DownloadAsset(asset, tempArchive, repo, release.TagName); err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}
	defer os.Remove(tempArchive)

	fmt.Printf("Extracting %s...\n", commandName)

	opts := archive.ExtractionOptions{
		DestDir:    cmdDir,
		FileFilter: archive.DefaultFileFilter(commandName),
	}

	if err := archive.ExtractTarGz(tempArchive, opts); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	binaryPath := filepath.Join(cmdDir, commandName)
	if _, err := os.Stat(binaryPath); err != nil {
		return fmt.Errorf("extracted binary not found at %s", binaryPath)
	}

	if err := os.Chmod(binaryPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	if err := version.WriteReleaseInfo(cmdDir, commandName, release.TagName, repo); err != nil {
		return fmt.Errorf("failed to write version info: %w", err)
	}

	fmt.Printf("Successfully installed %s %s\n", commandName, release.TagName)
	return nil
}

func getStateHome() string {
	stateHome := os.Getenv("XDG_STATE_HOME")
	if stateHome == "" {
		homeDir, _ := os.UserHomeDir()
		stateHome = filepath.Join(homeDir, ".local", "state")
	}
	return stateHome
}