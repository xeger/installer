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
	repoPatterns := github.FindRepositories(commandName)

	var release *github.Release
	var repo string
	var err error

	for _, pattern := range repoPatterns {
		release, err = github.GetLatestRelease(pattern)
		if err == nil {
			repo = pattern
			break
		}
	}

	if release == nil {
		return fmt.Errorf(github.FormatRepositoryNotFoundError(commandName, repoPatterns))
	}

	fmt.Printf("Installing %s from %s@%s...\n", commandName, repo, release.TagName)

	asset, err := github.FindAssetForPlatform(release, commandName)
	if err != nil {
		return fmt.Errorf("no compatible release asset: %w", err)
	}

	fmt.Printf("Downloading %s...\n", asset.Name)

	dataHome := getDataHome()
	cmdDir := filepath.Join(dataHome, "crossnokaye", "cli", "cmd", commandName)
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

func getDataHome() string {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		homeDir, _ := os.UserHomeDir()
		dataHome = filepath.Join(homeDir, ".local", "share")
	}
	return dataHome
}
