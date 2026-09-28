// Package updates decides when to look for new releases and whether a
// release is newer than what is installed.
package updates

import (
	"time"

	"golang.org/x/mod/semver"
)

// Interval is how often the launcher looks for new releases of a tool, or of itself.
const Interval = 7 * 24 * time.Hour

// Due reports whether a check last made at last should be repeated at now.
func Due(last, now time.Time) bool {
	return now.Sub(last) >= Interval
}

// Newer reports whether latest supersedes current. Semantic versions are
// compared as such; anything else counts as newer whenever it differs.
func Newer(current, latest string) bool {
	if semver.IsValid(current) && semver.IsValid(latest) {
		return semver.Compare(latest, current) > 0
	}
	return latest != "" && latest != current
}

// IsRelease reports whether version is a plain release tag (e.g. v1.2.3)
// rather than a dev build, pseudo-version, or dirty build.
func IsRelease(version string) bool {
	return semver.IsValid(version) && semver.Prerelease(version) == "" && semver.Build(version) == ""
}
