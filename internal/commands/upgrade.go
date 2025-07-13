package commands

import (
	"fmt"
	"path/filepath"

	"github.com/crossnokaye/cli/internal/github"
	"github.com/crossnokaye/cli/internal/version"
)

func Upgrade(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: ck upgrade <command>")
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

