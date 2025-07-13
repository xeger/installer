package commands

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/crossnokaye/cli/internal/archive"
	"github.com/crossnokaye/cli/internal/github"
)

func Upgrade(args []string) error {
	if len(args) == 0 {
		return upgradeSelf()
	}

	commandName := args[0]
	fmt.Printf("Upgrading %s...\n", commandName)

	return Install([]string{commandName})
}

func upgradeSelf() error {
	fmt.Println("Upgrading ck...")

	repo := "crossnokaye/cli"
	commandName := "ck"

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