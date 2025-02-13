package version

import (
	"fmt"
	"os/exec"
	"runtime/debug"
	"strings"
)

// These variables are populated by the build system
var (
	// Version is the current version of the application
	Version = "dev"
	// GitCommit is the git commit hash
	GitCommit = ""
	// GitBranch is the git branch
	GitBranch = ""
)

// GetVersion returns the full version string including git details for development versions
func GetVersion() string {
	if Version != "dev" {
		return Version
	}

	// Try to get version from build info first
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	// For development builds, include git details
	commit := getGitCommit()
	branch := getGitBranch()
	if commit == "" && branch == "" {
		return "dev"
	}

	version := "dev"
	if branch != "" {
		version += fmt.Sprintf("-%s", branch)
	}
	if commit != "" {
		version += fmt.Sprintf("-%s", commit[:8]) // Use first 8 chars of commit hash
	}
	return version
}

// getGitCommit returns the current git commit hash
func getGitCommit() string {
	if GitCommit != "" {
		return GitCommit
	}
	cmd := exec.Command("git", "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// getGitBranch returns the current git branch
func getGitBranch() string {
	if GitBranch != "" {
		return GitBranch
	}
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
