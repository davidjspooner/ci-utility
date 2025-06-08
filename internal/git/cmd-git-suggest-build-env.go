package git

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

// SuggestBuildEnvOptions holds options for the suggest-build-env command.
type SuggestBuildEnvOptions struct {
	CommandPrefix string `flag:"--command-prefix,Prefix for command output"`
	VersionPrefix string `flag:"--version-prefix,Prefix string for version"`
	VersionSuffix string `flag:"--version-suffix,Suffix string for version"`
}

// executeSuggestBuildEnv prints environment variables for the current build context.
// It determines the new tag, build branch, version, and context, and prints them as shell variables.
func executeSuggestBuildEnv(ctx context.Context, options *SuggestBuildEnvOptions, args []string) error {

	// Add # prefix to logging.
	handler := slog.Default().Handler()
	if handler != nil {
		cmdLogHandler, ok := handler.(*cmd.LogHandler)
		if ok {
			cmdLogHandler.Prefix = "# "
		}
	}

	// Get the current branch.
	currentBranch, err := GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %v", err)
	}

	// Get the latest tag and version.
	latestTag, currentVersion, err := getLatestTagAndVersion(ctx, currentBranch)
	if err != nil {
		return fmt.Errorf("failed to get the latest tag: %v", err)
	}

	// Log the current tag and version.
	slog.DebugContext(ctx, "Current", "tag", latestTag, "version", currentVersion.String())

	prefix := options.CommandPrefix

	now := time.Now().UTC()
	nowStr := now.Format(time.RFC1123)
	// Print environment variables for the build.
	fmt.Printf("%sBUILD_BRANCH=%s\n", prefix, quoteIfNeeded(currentBranch))
	fmt.Printf("%sBUILD_VERSION=%s\n", prefix, quoteIfNeeded(suggestBuildName()))
	fmt.Printf("%sBUILD_FROM=%s\n", prefix, quoteIfNeeded(getGitUrl()))
	fmt.Printf("%sBUILD_BY=%s\n", prefix, quoteIfNeeded(getBuildContext()))
	fmt.Printf("%sBUILD_TIME=%s\n", prefix, quoteIfNeeded(nowStr))
	return nil
}

func quoteIfNeeded(s string) string {
	if strings.ContainsAny(s, " \t\n") {
		return fmt.Sprintf("%q", s)
	}
	return s
}

// suggestBuildName returns a string representing the build version or identifier.
// It checks for uncommitted changes, tags, or falls back to the commit hash.
func suggestBuildName() string {
	// Check for uncommitted changes.
	out, err := Run("status", "--porcelain")
	if err != nil {
		return "UNKNOWN"
	}
	if strings.TrimSpace(out) != "" {
		// If there are uncommitted changes, use a timestamp.
		return "HEAD." + time.Now().Format("060102.1504")
	}

	// Check for a tag version.
	tag, err := Run("tag", "--contains", "HEAD")
	if err == nil && tag != "" {
		tags := strings.Split(strings.TrimSpace(tag), "\n")
		tag = strings.TrimSpace(tags[0]) // Take the first tag if multiple are returned.
		// If HEAD is tagged, return the tag.
		return tag
	}

	// Fallback to short commit hash.
	commitHash, err := Run("rev-parse", "--short", "HEAD")
	if err != nil {
		return "UNKNOWN"
	}
	return "COMMIT." + commitHash
}

// getBuildContext returns a string describing the build context, such as a CI run ID or user@hostname.
func getBuildContext() string {
	// Check for github actions.
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return os.ExpandEnv("${GITHUB_JOB} (${GITHUB_RUN_ID}@github)")
	}

	// Fallback to user@hostname.
	user := os.Getenv("USER")
	if user == "" {
		user = "unknown"
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return user + "@" + hostname
}

func getGitUrl() string {
	// Get the remote URL for the origin.
	out, err := Run("config", "--get", "remote.origin.url")
	if err != nil {
		return "UNKNOWN"
	}
	out = strings.TrimSpace(out)

	// If the URL is an SSH URL, convert it to HTTPS.
	if strings.HasPrefix(out, "git@") {
		out = strings.ReplaceAll(out, ":", "/")
		out = strings.ReplaceAll(out, "git@", "https://")
	}

	out = strings.TrimSuffix(out, ".git")

	return out
}
