package github

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

// ReleaseCreateOptions holds the options for creating a GitHub release.
// It includes fields for the tag name, release name, body, and flags for draft and prerelease.
type ReleaseCreateOptions struct {
	TagName    string `flag:"--tag,Tag name for the release"`
	Name       string `flag:"--name|--title,Human name of the release (defaults to the tag name)"`
	Body       string `flag:"--body,Description of the release"`
	Draft      bool   `flag:"--draft,Create the release as a draft"`
	Prerelease bool   `flag:"--prerelease,Mark the release as a prerelease"`
}

// executeGithubReleaseCreate creates a GitHub release and uploads assets.
func executeGithubReleaseCreate(ctx context.Context, option *ReleaseCreateOptions, args []string) error {
	// Validate the required options.
	files, err := globFiles(args)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no files found matching the pattern")
	}

	// get the GitHub token and repository from environment variables.
	token := os.Getenv("GITHUB_TOKEN")
	repo := os.Getenv("GITHUB_REPOSITORY") // e.g., "owner/repo"

	if token == "" || repo == "" {
		return fmt.Errorf("GITHUB_TOKEN and GITHUB_REPOSITORY environment variables are required")
	}

	// Create a GitHub API client with the provided token and repository information.
	client := Client{
		HTTPClient: http.DefaultClient,
		BaseURL:    "https://api.github.com",
		Token:      token,
		Owner:      strings.Split(repo, "/")[0],
		Repo:       strings.SplitN(repo, "/", 2)[1],
	}

	if option.TagName == "" {
		return fmt.Errorf("tag name (--tag) is required")
	}
	if option.Name == "" {
		option.Name = option.TagName // Default to tag name if not provided.
	}

	// prepare the release request payload.
	releaseReq := CreateReleaseRequest{
		TagName:    option.TagName,
		Name:       option.Name,
		Body:       option.Body,
		Draft:      option.Draft,
		Prerelease: option.Prerelease,
	}

	// Create the release using the GitHub API.
	var release ReleaseResponse
	err = client.PostJSON(ctx, "/repos/{owner}/{repo}/releases", releaseReq, &release)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create release", "tag", option.TagName, "error", err)
		os.Exit(1)
	}

	slog.InfoContext(ctx, "Created release", "id", release.ID, "tag", option.TagName, "name", release.Name, "url", release.URL)

	// Upload each file as an asset to the created release.
	for _, path := range files {
		var asset AssetResponse
		err = client.UploadBinaryFile(ctx, release.ID, path, &asset)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to upload asset", "name", path, "error", err)
			os.Exit(1)
		}
		slog.InfoContext(ctx, "Uploaded asset", "name", asset.Name, "url", asset.URL)
	}

	return nil
}
