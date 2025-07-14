package commands

import (
	"fmt"
	"path/filepath"

	"github.com/crossnokaye/cli/internal/github"
	"github.com/crossnokaye/cli/internal/version"
)

func Upgrade(args []string) error {
	if len(args) == 0 {
		return upgradeAll()
	}

	var errors []string
	successCount := 0

	for _, commandName := range args {
		if err := upgradeOne(commandName); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", commandName, err))
		} else {
			successCount++
		}
	}

	return printUpgradeSummary(successCount, len(args), errors)
}

// upgradeAll upgrades all installed commands
func upgradeAll() error {
	dataHome := getDataHome()

	// Get all installed commands
	installedVersions, err := version.GetAllInstalledVersions(dataHome)
	if err != nil {
		return fmt.Errorf("failed to get installed commands: %w", err)
	}

	if len(installedVersions) == 0 {
		fmt.Println("No installed commands found to upgrade.")
		return nil
	}

	fmt.Printf("Upgrading %d installed commands...\n", len(installedVersions))

	var errors []string
	successCount := 0

	for commandName := range installedVersions {
		fmt.Printf("\n--- Upgrading %s ---\n", commandName)
		if err := upgradeOne(commandName); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", commandName, err))
		} else {
			successCount++
		}
	}

	return printUpgradeSummary(successCount, len(installedVersions), errors)
}

// upgradeOne upgrades a single command
func upgradeOne(commandName string) error {
	fmt.Printf("Checking for %s upgrades...\n", commandName)

	// Check current version
	dataHome := getDataHome()
	cmdDir := filepath.Join(dataHome, "crossnokaye", "cli", "cmd", commandName)

	currentVersion, err := version.GetInstalledVersion(cmdDir, commandName)
	if err != nil {
		fmt.Printf("No current version found for %s, installing...\n", commandName)
		return Install([]string{commandName})
	}

	// Find our target repository
	repoPatterns := github.FindRepositories(commandName)

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

// printUpgradeSummary prints the upgrade results and returns an error if any upgrades failed
func printUpgradeSummary(successCount, totalCount int, errors []string) error {
	fmt.Printf("\n--- Upgrade Summary ---\n")
	fmt.Printf("Successfully upgraded: %d/%d commands\n", successCount, totalCount)

	if len(errors) > 0 {
		fmt.Printf("Failed to upgrade:\n")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
		return fmt.Errorf("failed to upgrade %d commands", len(errors))
	}

	return nil
}
