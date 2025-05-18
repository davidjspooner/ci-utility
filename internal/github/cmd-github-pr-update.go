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

type GithubPRUpdateOptions struct {
	PRNumber string `flag:"<pr-number>,Pull request number"`
	DryRun   bool   `flag:"--dry-run,Do not update the PR title"`
}

func executeUpdateGithubPRMeta(ctx context.Context, option *GithubPRUpdateOptions, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: update-github-pr-meta <pr-number> [--dry-run]")
	}

	if option.PRNumber == "" {
		return fmt.Errorf("pull request number is required")
	}

	token := os.Getenv("GITHUB_TOKEN")
	repo := os.Getenv("GITHUB_REPOSITORY") // e.g., "owner/repo"

	if token == "" || repo == "" {
		return fmt.Errorf("GITHUB_TOKEN and GITHUB_REPOSITORY environment variables are required")
	}

	// Fetch commit messages
	commitMessages := getCommitMessages(option.PRNumber, token, repo)

	// Determine version bump
	bump, err := semantic.Bumps.GetVersionBump(commitMessages)
	if err != nil {
		return fmt.Errorf("error determining bump : %v", err)
	}

	//get the current PR title
	prTitle, err := getPullRequestTitle(ctx, option.PRNumber, token, repo)
	if err != nil {
		return fmt.Errorf("error fetching PR title: %v", err)
	}
	if strings.Contains(prTitle, bump) {
		slog.Info("PR title already contains the bump", "pr", option.PRNumber, "bump", bump)
		return nil
	}

	// Compose new title
	newTitle := fmt.Sprintf("%s: update based on commits", bump)

	if option.DryRun {
		slog.Warn("--dry-run", "pr", option.PRNumber, "title", newTitle)
		return nil
	}

	// Update PR via GitHub API
	return updatePullRequest(ctx, option.PRNumber, token, repo, newTitle)
}

func getCommitMessages(prNumber, token, repo string) []string {
	url := fmt.Sprintf("https://api.github.com/repos/%s/pulls/%s/commits", repo, prNumber)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		log.Fatalf("Failed to fetch commits: %v", err)
	}
	defer resp.Body.Close()

	var result []struct {
		Message string `json:"commit.message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("Decode error: %v", err)
	}

	var messages []string
	for _, c := range result {
		messages = append(messages, c.Message)
	}
	return messages
}

func updatePullRequest(_ context.Context, prNumber, token, repo, title string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/pulls/%s", repo, prNumber)

	pr := struct {
		Title string `json:"title"`
	}{Title: title}
	data, _ := json.Marshal(pr)

	req, _ := http.NewRequest("PATCH", url, strings.NewReader(string(data)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to update PR: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to update PR, status code: %d", resp.StatusCode)
	}

	slog.Info("Updated PR title", "pr", prNumber, "title", title)
	return nil
}

func getPullRequestTitle(_ context.Context, prNumber, token, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/pulls/%s", repo, prNumber)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to fetch PR title: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to fetch PR title, status code: %d", resp.StatusCode)
	}

	var result struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode error: %v", err)
	}

	return result.Title, nil
}
