package github

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type GithubReleaseCreateOptions struct {
	TagName    string `flag:"--tag,Tag name for the release"`
	Name       string `flag:"--name|--title,Human name of the release (defaults to the tag name)"`
	Body       string `flag:"--body,Description of the release"`
	Draft      bool   `flag:"--draft,Create the release as a draft"`
	Prerelease bool   `flag:"--prerelease,Mark the release as a prerelease"`
}

func executeGithubReleaseCreate(ctx context.Context, option *GithubReleaseCreateOptions, args []string) error {

	files, err := globFiles(args)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no files found matching the pattern")
	}

	token := os.Getenv("GITHUB_TOKEN")
	repo := os.Getenv("GITHUB_REPOSITORY") // e.g., "owner/repo"

	if token == "" || repo == "" {
		return fmt.Errorf("GITHUB_TOKEN and GITHUB_REPOSITORY environment variables are required")
	}

	client := Client{
		HTTPClient: http.DefaultClient,
		BaseURL:    "https://api.github.com",
		Token:      token,
		Owner:      strings.Split(repo, "/")[0],
		Repo:       strings.SplitN(repo, "/", 2)[1],
	}

	releaseReq := CreateReleaseRequest{
		TagName:    option.TagName,
		Name:       option.Name,
		Body:       option.Body,
		Draft:      option.Draft,
		Prerelease: option.Prerelease,
	}

	var release ReleaseResponse
	err = client.PostJSON(ctx, "/repos/{owner}/{repo}/releases", releaseReq, &release)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create release", "tag", option.TagName, "error", err)
		os.Exit(1)
	}

	slog.InfoContext(ctx, "Created release", "id", release.ID, "tag", option.TagName, "name", release.Name, "url", release.URL)

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
