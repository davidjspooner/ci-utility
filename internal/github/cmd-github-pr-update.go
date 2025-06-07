package github

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/davidjspooner/ci-utility/pkg/semantic"
)

// PRUpdateOptions holds options for updating a GitHub PR.
type PRUpdateOptions struct {
	// PRNumber is the pull request number.
	PRNumber string `flag:"<pr-number>,Pull request number"`
	// DryRun indicates whether to perform a dry run (no actual updates).
	DryRun bool `flag:"--dry-run,Do not update the PR title"`
}

// executeUpdateGithubPRMeta updates the metadata (title) of a GitHub PR based on commit messages.
func executeUpdateGithubPRMeta(ctx context.Context, option *PRUpdateOptions, args []string) error {
	// Check for required arguments.
	if len(args) < 1 {
		return fmt.Errorf("usage: update-github-pr-meta <pr-number> [--dry-run]")
	}

	// Ensure PR number is provided.
	if option.PRNumber == "" {
		return fmt.Errorf("pull request number is required")
	}

	// Read GitHub token and repository from environment.
	token := os.Getenv("GITHUB_TOKEN")
	repo := os.Getenv("GITHUB_REPOSITORY") // e.g., "owner/repo"

	// Validate required environment variables.
	if token == "" || repo == "" {
		return fmt.Errorf("GITHUB_TOKEN and GITHUB_REPOSITORY environment variables are required")
	}

	// Fetch commit messages for the PR.
	commitMessages := getCommitMessages(option.PRNumber, token, repo)

	// Determine the semantic version bump from commit messages.
	bump, reason, err := semantic.Bumps.GetVersionBump(commitMessages)
	if err != nil {
		return fmt.Errorf("error determining bump : %v", err)
	}

	slog.Debug("Found commit which needs version increament", "level", bump, "message", reason)

	// Get the current PR title.
	prTitle, err := getPullRequestTitle(ctx, option.PRNumber, token, repo)
	if err != nil {
		return fmt.Errorf("error fetching PR title: %v", err)
	}
	// Check if the bump is already present in the title.
	if strings.Contains(prTitle, bump) {
		slog.InfoContext(ctx, "PR title already contains the bump", "pr", option.PRNumber, "bump", bump)
		return nil
	}

	// Compose the new PR title.
	newTitle := fmt.Sprintf("%s: update based on commits", bump)

	// If dry run, log and exit.
	if option.DryRun {
		slog.WarnContext(ctx, "--dry-run", "pr", option.PRNumber, "title", newTitle)
		return nil
	}

	// Update the PR title via GitHub API.
	return updatePullRequest(ctx, option.PRNumber, token, repo, newTitle)
}

// getCommitMessages fetches commit messages for a given PR from GitHub.
func getCommitMessages(prNumber, token, repo string) []string {
	url := fmt.Sprintf("https://api.github.com/repos/%s/pulls/%s/commits", repo, prNumber)

	// Prepare the HTTP request.
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	// Check for errors in the HTTP response.
	if err != nil || resp.StatusCode != 200 {
		log.Fatalf("Failed to fetch commits: %v", err)
	}
	defer resp.Body.Close()

	// Parse the JSON response.
	var result []struct {
		Message string `json:"commit.message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("Decode error: %v", err)
	}

	// Extract commit messages.
	var messages []string
	for _, c := range result {
		messages = append(messages, c.Message)
	}
	return messages
}

// updatePullRequest updates the title of a GitHub PR.
func updatePullRequest(ctx context.Context, prNumber, token, repo, title string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/pulls/%s", repo, prNumber)

	// Prepare the request body.
	pr := struct {
		Title string `json:"title"`
	}{Title: title}
	data, _ := json.Marshal(pr)

	// Create the HTTP PATCH request.
	req, _ := http.NewRequest("PATCH", url, strings.NewReader(string(data)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	// Check for errors in the HTTP response.
	if err != nil || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to update PR: %v", err)
	}
	defer resp.Body.Close()

	// Ensure the status code is 200 OK.
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to update PR, status code: %d", resp.StatusCode)
	}

	// Log the successful update.
	slog.InfoContext(ctx, "Updated PR title", "pr", prNumber, "title", title)
	return nil
}

// getPullRequestTitle fetches the title of a GitHub PR.
func getPullRequestTitle(_ context.Context, prNumber, token, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/pulls/%s", repo, prNumber)

	// Prepare the HTTP request.
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	// Check for errors in the HTTP response.
	if err != nil || resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to fetch PR title: %v", err)
	}
	defer resp.Body.Close()
	// Ensure the status code is 200 OK.
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to fetch PR title, status code: %d", resp.StatusCode)
	}

	// Parse the JSON response.
	var result struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode error: %v", err)
	}

	// Return the PR title.
	return result.Title, nil
}
