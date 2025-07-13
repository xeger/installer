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
