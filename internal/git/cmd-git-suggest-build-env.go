package git

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/davidjspooner/ci-utility/pkg/semantic"
	"github.com/davidjspooner/go-text/pkg/cmd"
)

type GetGitEnvOptions struct {
	VersionPrefix string `flag:"--version-prefix,Prefix string for version"`
	VersionSuffix string `flag:"--version-suffix,Suffix string for version"`
}

func executeGetGitEnv(ctx context.Context, options *GetGitEnvOptions, args []string) error {

	//add # prefix to logging
	handler := slog.Default().Handler()
	if handler != nil {
		cmdLogHandler, ok := handler.(*cmd.LogHandler)
		if ok {
			cmdLogHandler.Prefix = "# "
		}
	}

	// Get the current branch
	currentBranch, err := GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %v", err)
	}

	// Get the latest tag
	latestTag, currentVersion, err := getLatestTagAndVersion(ctx, currentBranch)
	if err != nil {
		return fmt.Errorf("failed to get the latest tag: %v", err)
	}

	slog.Info("Current", "tag", latestTag, "version", currentVersion.String())

	// Get commit messages since the latest tag
	validCommitMessages, commits, err := getCommitsSinceTag(latestTag)
	if err != nil {
		return err
	}

	newTag := ""
	if len(validCommitMessages) > 0 {

		// Determine the version reason
		reason, err := semantic.Bumps.GetVersionBump(commits)
		if err != nil {
			return fmt.Errorf("failed to determine version increment: %v", err)
		}

		// Generate the new tag
		newTag, err = generateNewTag(options.VersionPrefix, options.VersionSuffix, currentVersion, reason)
		if err != nil {
			return err
		}
	}

	fmt.Printf("NEW_TAG=%s\n", newTag)
	fmt.Printf("BUILD_BRANCH=%s\n", currentBranch)
	fmt.Printf("BUILD_VERSION=%s\n", suggestBuildName())
	fmt.Printf("BUILD_CONTEXT=%s\n", getBuildContext())
	now := time.Now().UTC()
	fmt.Printf("BUILD_TIME=%s\n", now.Format(time.RFC1123))
	return nil
}

func suggestBuildName() string {
	// Check for uncommitted changes
	out, err := Run("status", "--porcelain")
	if err != nil {
		return "UNKNOWN"
	}
	if strings.TrimSpace(out) != "" {
		return "HEAD." + time.Now().Format("060102.1504")
	}

	// Check for a tag version
	tag, err := Run("tag", "--contains", "HEAD")
	if err == nil && tag != "" {
		return tag
	}

	// Fallback to short commit hash
	commitHash, err := Run("rev-parse", "--short", "HEAD")
	if err != nil {
		return "UNKNOWN"
	}
	return commitHash
}

func getBuildContext() string {
	// Check for github actions
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return os.Getenv("GITHUB_RUN_ID")
	}

	// Fallback to user@hostname
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
