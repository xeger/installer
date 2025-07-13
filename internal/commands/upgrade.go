package commands

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/crossnokaye/cli/internal/archive"
	"github.com/crossnokaye/cli/internal/github"
	"github.com/crossnokaye/cli/internal/version"
)

func Upgrade(args []string) error {
	if len(args) == 0 {
		return upgradeSelf()
	}

	commandName := args[0]
	fmt.Printf("Checking for %s upgrades...\n", commandName)

	// Check current version
	stateHome := getStateHome()
	cmdDir := filepath.Join(stateHome, "crossnokaye", "cli", "cmd", commandName)

	currentVersion, err := version.GetInstalledVersion(cmdDir, commandName)
	if err != nil {
		fmt.Printf("No current version found for %s, installing...\n", commandName)
		return Install([]string{commandName})
	}

	// Get latest version
	repoPatterns := []string{
		fmt.Sprintf("crossnokaye/cli-%s", commandName),
		fmt.Sprintf("crossnokaye/%s", commandName),
	}

	var release *github.Release

	for _, pattern := range repoPatterns {
		release, err = github.GetLatestRelease(pattern)
		if err == nil {
			break
		}
	}

	if release == nil {
		return fmt.Errorf("failed to find repository for %s", commandName)
	}

	if currentVersion == release.TagName {
		fmt.Printf("%s is already up to date (%s)\n", commandName, currentVersion)
		return nil
	}

	fmt.Printf("Upgrading %s from %s to %s...\n", commandName, currentVersion, release.TagName)
	return Install([]string{commandName})
}

func upgradeSelf() error {
	fmt.Println("Checking for ck upgrades...")

	repo := "crossnokaye/cli"
	commandName := "ck"

	// Check current version if possible
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		if currentVersion, err := version.GetInstalledVersion(execDir, commandName); err == nil {
			release, err := github.GetLatestRelease(repo)
			if err == nil && currentVersion == release.TagName {
				fmt.Printf("ck is already up to date (%s)\n", currentVersion)
				return nil
			}
		}
	}

	fmt.Println("Upgrading ck...")

	release, err := github.GetLatestRelease(repo)
	if err != nil {
		return fmt.Errorf("failed to get latest release: %w", err)
	}

	fmt.Printf("Found release %s\n", release.TagName)

	asset, err := github.FindAssetForPlatform(release, commandName)
	if err != nil {
		return fmt.Errorf("failed to find compatible asset: %w", err)
	}

	fmt.Printf("Downloading %s...\n", asset.Name)

	tempDir, err := os.MkdirTemp("", "ck-upgrade-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	tempArchive := filepath.Join(tempDir, asset.Name)
	if err := github.DownloadAsset(asset, tempArchive, repo, release.TagName); err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}

	fmt.Printf("Extracting and testing new binary...\n")

	tempBinary := filepath.Join(tempDir, commandName)
	opts := archive.ExtractionOptions{
		DestDir:    tempDir,
		FileFilter: archive.DefaultFileFilter(commandName),
	}

	if err := archive.ExtractTarGz(tempArchive, opts); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	if _, err := os.Stat(tempBinary); err != nil {
		return fmt.Errorf("extracted binary not found at %s", tempBinary)
	}

	if err := os.Chmod(tempBinary, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	testCmd := exec.Command(tempBinary, "help")
	if err := testCmd.Run(); err != nil {
		return fmt.Errorf("new binary failed help test: %w", err)
	}

	fmt.Printf("New binary tested successfully. Installing...\n")

	currentBinary, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	if err := replaceBinary(tempBinary, currentBinary); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	// Write version info for ck - use executable directory
	execDir := filepath.Dir(currentBinary)
	if err := version.WriteReleaseInfo(execDir, "ck", release.TagName, repo); err != nil {
		// Non-fatal - just log the error
		fmt.Printf("Warning: failed to write version info: %v\n", err)
	}

	fmt.Printf("Successfully upgraded ck to %s\n", release.TagName)
	return nil
}

func replaceBinary(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source binary: %w", err)
	}
	defer srcFile.Close()

	tempDest := dest + ".new"
	destFile, err := os.OpenFile(tempDest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create new binary: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		os.Remove(tempDest)
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	destFile.Close()

	if err := os.Rename(tempDest, dest); err != nil {
		os.Remove(tempDest)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return nil
}
