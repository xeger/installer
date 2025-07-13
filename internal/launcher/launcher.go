package launcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/crossnokaye/cli/internal/version"
)

func Execute(subcommand string, args []string) error {
	binaryPath, err := findBinary(subcommand)
	if err != nil {
		return err
	}

	cmd := exec.Command(binaryPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func findBinary(subcommand string) (string, error) {
	dataHome := getDataHome()
	binaryPath := filepath.Join(dataHome, "crossnokaye", "cli", "cmd", subcommand, subcommand)

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return "", fmt.Errorf("subcommand '%s' not found at %s", subcommand, binaryPath)
	}

	return binaryPath, nil
}

func getDataHome() string {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		homeDir, _ := os.UserHomeDir()
		dataHome = filepath.Join(homeDir, ".local", "share")
	}
	return dataHome
}

func GetAvailableCommands() ([]string, error) {
	dataHome := getDataHome()
	cmdDir := filepath.Join(dataHome, "crossnokaye", "cli", "cmd")

	if _, err := os.Stat(cmdDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read commands directory: %w", err)
	}

	var commands []string
	for _, entry := range entries {
		if entry.IsDir() {
			binaryPath := filepath.Join(cmdDir, entry.Name(), entry.Name())
			if _, err := os.Stat(binaryPath); err == nil {
				commands = append(commands, entry.Name())
			}
		}
	}

	sort.Strings(commands)
	return commands, nil
}

func PrintUsage(programName string) {
	programName = filepath.Base(programName)

	dataHome := getDataHome()
	commands, versions, err := version.GetSortedCommandsWithVersions(dataHome)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing commands: %v\n", err)
		commands = []string{}
		versions = make(map[string]string)
	}

	fmt.Printf("Usage: %s <subcommand> [args...]\n\n", programName)

	fmt.Printf("CrossnoKaye meta-CLI and version manager\n\n")

	fmt.Printf("Built-in subcommands:\n")
	fmt.Printf("  install   Install a subcommand from GitHub releases\n")
	fmt.Printf("  upgrade   Upgrade a subcommand to the latest version\n")

	if len(commands) > 0 {
		fmt.Printf("\nInstalled subcommands:\n")
		for _, cmd := range commands {
			if ver, exists := versions[cmd]; exists {
				fmt.Printf("  %-16s %s\n", cmd, ver)
			} else {
				fmt.Printf("  %-16s %s\n", cmd, "unknown")
			}
		}
	} else {
		fmt.Printf("\nNo installed subcommands found.\n")
		fmt.Printf("Use '%s install <command>' to install subcommands from GitHub.\n", programName)
	}

	fmt.Printf("\nOptions:\n")
	fmt.Printf("  --help    Show this help message\n")
	fmt.Printf("  help      Show this help message\n")
}
