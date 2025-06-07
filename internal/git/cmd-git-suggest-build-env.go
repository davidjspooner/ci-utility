package git

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/davidjspooner/ci-utility/pkg/semantic"
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

	// Get commit messages since the latest tag.
	validCommitMessages, commits, err := getCommitsSinceTag(latestTag)
	if err != nil {
		return err
	}

	// Prepare newTag variable.
	newTag := ""
	if len(validCommitMessages) > 0 {

		// Determine the version reason.
		reason, err := semantic.Bumps.GetVersionBump(commits)
		if err != nil {
			return fmt.Errorf("failed to determine version increment: %v", err)
		}

		// Generate the new tag.
		newTag, err = generateNewTag(options.VersionPrefix, options.VersionSuffix, currentVersion, reason)
		if err != nil {
			return err
		}
	}

	prefix := options.CommandPrefix
	if prefix != "" && !strings.HasSuffix(prefix, " ") {
		prefix += " "
	}

	// Print environment variables for the build.
	fmt.Printf("%sNEW_TAG=%q\n", prefix, newTag)
	fmt.Printf("%sBUILD_BRANCH=%q\n", prefix, currentBranch)
	fmt.Printf("%sBUILD_VERSION=%q\n", prefix, suggestBuildName())
	fmt.Printf("%sBUILD_BY=%q\n", prefix, getBuildContext())
	now := time.Now().UTC()
	fmt.Printf("%sBUILD_TIME=%q\n", prefix, now.Format(time.RFC1123))
	return nil
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
		// If HEAD is tagged, return the tag.
		return tag
	}

	// Fallback to short commit hash.
	commitHash, err := Run("rev-parse", "--short", "HEAD")
	if err != nil {
		return "UNKNOWN"
	}
	return commitHash
}

// getBuildContext returns a string describing the build context, such as a CI run ID or user@hostname.
func getBuildContext() string {
	// Check for github actions.
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return os.Getenv("GITHUB_RUN_ID")
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
