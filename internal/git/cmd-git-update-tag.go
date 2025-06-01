package git

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/davidjspooner/ci-utility/pkg/semantic"
)

// BumpGitTagOptions holds options for bumping git tags.
type BumpGitTagOptions struct {
	Prefix string `flag:"--prefix,Prefix string"`
	Suffix string `flag:"--suffix,Suffix string"`
	DryRun bool   `flag:"--dry-run,Do not push the tag"`
	Remote string `flag:"--remote,Remote to push the tag to"`
}

// generateNewTag creates a new tag string based on the prefix, suffix, current version, and bump reason.
func generateNewTag(prefix, suffix string, currentVersion semantic.Version, reason string) (string, error) {
	// Increment the version
	newVersion, err := currentVersion.Increment(reason)
	if err != nil {
		return "", fmt.Errorf("failed to increment version: %v", err)
	}

	// Construct the new tag
	newTag := fmt.Sprintf("%s%s%s", prefix, newVersion.String(), suffix)
	return newTag, nil
}

// applyNewTag creates and pushes the new tag, unless DryRun is set.
func applyNewTag(ctx context.Context, newTag string, option *BumpGitTagOptions) error {
	if option.DryRun {
		slog.InfoContext(ctx, "--dry-run", "newTag", newTag)
		return nil
	}

	// Create and push the new tag
	if _, err := Run("tag", newTag); err != nil {
		return fmt.Errorf("failed to create tag: %v", err)
	}
	if _, err := Run("push", option.Remote, newTag); err != nil {
		return fmt.Errorf("failed to push tag: %v", err)
	}

	slog.InfoContext(ctx, "Tag created and pushed", "tag", newTag)
	return nil
}

// executeBumpGitTag determines the next version and applies a new tag based on commit messages.
func executeBumpGitTag(ctx context.Context, option *BumpGitTagOptions, args []string) error {

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

	slog.InfoContext(ctx, "Current", "tag", latestTag, "version", currentVersion.String())

	// Get commit messages since the latest tag
	validCommitMessages, commits, err := getCommitsSinceTag(latestTag)
	if err != nil {
		return err
	}
	if len(validCommitMessages) == 0 {
		fmt.Println("No changes detected, no version increment needed.")
		return nil
	}

	// Determine the version reason
	reason, err := semantic.Bumps.GetVersionBump(commits)
	if err != nil {
		return fmt.Errorf("failed to determine version increment: %v", err)
	}

	// Generate the new tag
	newTag, err := generateNewTag(option.Prefix, option.Suffix, currentVersion, reason)
	if err != nil {
		return err
	}

	slog.DebugContext(ctx, "Increment", "reason", reason)

	// Apply the new tag
	return applyNewTag(ctx, newTag, option)
}

// getCommitsSinceTag returns commit messages and raw commits since the given tag.
func getCommitsSinceTag(latestTag string) ([]string, []string, error) {
	commitMessages, err := Run("log", fmt.Sprintf("%s..HEAD", latestTag), "--pretty=format:%s")
	validCommitMessages := []string{}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get commit messages: %v", err)
	}
	commits := splitLines(commitMessages)
	for _, commit := range commits {
		commit = strings.TrimSpace(commit)
		if commit != "" {
			validCommitMessages = append(validCommitMessages, commit)
		}
	}
	return validCommitMessages, commits, nil
}

// getLatestTagAndVersion finds the latest tag and its semantic version for the given branch.
func getLatestTagAndVersion(ctx context.Context, branch string) (string, semantic.Version, error) {
	commits, err := Run("rev-list", "--tags", "--no-walk", "--abbrev=0", "--date-order", branch)
	if err != nil {
		return "", semantic.Version{}, fmt.Errorf("failed to get latest tags: %v", err)
	}
	var bestVersion semantic.Version
	var bestTag string

	slog.DebugContext(ctx, "Latest commits found", "commits", commits)
	commitList := splitLines(commits)
	for _, commit := range commitList {
		if commit == "" {
			continue
		}
		commit = strings.TrimSpace(commit)
		checkCmd, err := Run("merge-base", "--is-ancestor", commit, branch)
		if err != nil {
			continue
		}
		slog.DebugContext(ctx, "Checked commit is ancestor", "commit", commit, "branch", branch, "checkCmd", checkCmd)
		tagsForCommit, err := Run("tag", "--contains", commit)
		if err != nil {
			continue
		}
		slog.DebugContext(ctx, "Tags for commit found", "commit", commit, "tags", tagsForCommit)

		for _, tag := range splitLines(tagsForCommit) {
			slog.DebugContext(ctx, "Tag found", "tag", tag)
			_, _, version, err := semantic.ExtractVersionFromTag(tag)
			if err != nil {
				continue
			}
			if version.IsGreaterThan(bestVersion) {
				bestVersion = version
				bestTag = tag
			}
		}
		if bestVersion.IsNotEmpty() {
			return bestTag, bestVersion, nil
		}
	}
	return "", semantic.Version{}, fmt.Errorf("no valid tags found for branch %s", branch)
}
