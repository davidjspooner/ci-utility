package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type Client struct {
	HTTPClient *http.Client
	BaseURL    string // eg. "https://api.github.com/"
	Token      string
	Owner      string
	Repo       string
	slog       slog.Logger
}

type GitHubError struct {
	StatusCode int
	Message    string                   `json:"message"`
	Errors     []map[string]interface{} `json:"errors,omitempty"`
	DocsURL    string                   `json:"documentation_url,omitempty"`
}

func (e *GitHubError) Error() string {
	return fmt.Sprintf("GitHub API error (%d): %s", e.StatusCode, e.Message)
}

func (c *Client) IsLogEnabled(ctx context.Context, level slog.Level) bool {
	if c.slog.Handler() == nil {
		return false
	}
	return c.slog.Enabled(ctx, level)
}

func (c *Client) Do(ctx context.Context, method, fullURL string, body io.Reader, contentType string, response interface{}) error {
	fullURL = strings.ReplaceAll(fullURL, "{owner}", url.PathEscape(c.Owner))
	fullURL = strings.ReplaceAll(fullURL, "{repo}", url.PathEscape(c.Repo))
	if !strings.HasPrefix(fullURL, "http") {
		fullURL = path.Join(c.BaseURL, fullURL)
	}

	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var ghErr GitHubError
		ghErr.StatusCode = resp.StatusCode
		_ = json.NewDecoder(resp.Body).Decode(&ghErr)
		return &ghErr
	}

	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) DoJSON(ctx context.Context, method, path string, request interface{}, response interface{}) error {
	var bodyReader io.Reader
	var b []byte
	var err error
	if request != nil {
		b, err = json.Marshal(request)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}
	fullURL := c.BaseURL + path
	err = c.Do(ctx, method, fullURL, bodyReader, "application/json", response)
	if c.IsLogEnabled(ctx, slog.LevelDebug) {
		//TODO log the request and responce details
	} else if err != nil && c.IsLogEnabled(ctx, slog.LevelWarn) {
		// log the error if any
	}
	return err
}

func (c *Client) GetJSON(ctx context.Context, path string, response interface{}) error {
	return c.DoJSON(ctx, http.MethodGet, path, nil, response)
}

func (c *Client) PostJSON(ctx context.Context, path string, request interface{}, response interface{}) error {
	return c.DoJSON(ctx, http.MethodPost, path, request, response)
}

func (c *Client) PutJSON(ctx context.Context, path string, request interface{}, response interface{}) error {
	return c.DoJSON(ctx, http.MethodPut, path, request, response)
}

func (c *Client) DeleteJSON(ctx context.Context, path string, response interface{}) error {
	return c.DoJSON(ctx, http.MethodDelete, path, nil, response)
}

type UploadMeta struct {
	ReleaseID int64
	Name      string
	Label     string
	UploadURL string // Optional, if provided will be used instead of constructing the URL
}

func (c *Client) UploadBinary(ctx context.Context, meta UploadMeta, contentType string, data io.Reader, response interface{}) error {
	base := meta.UploadURL
	if base == "" {
		base = fmt.Sprintf("https://uploads.github.com/repos/%s/%s/releases/%d/assets", url.PathEscape(c.Owner), url.PathEscape(c.Repo), meta.ReleaseID)
	}

	u, err := url.Parse(base)
	if err != nil {
		return fmt.Errorf("failed to build upload URL: %w", err)
	}

	q := u.Query()
	q.Set("name", meta.Name)
	if meta.Label != "" {
		q.Set("label", meta.Label)
	}
	u.RawQuery = q.Encode()

	err = c.Do(ctx, http.MethodPost, u.String(), data, contentType, response)
	if c.IsLogEnabled(ctx, slog.LevelDebug) {
		//TODO log the request and responce details
	} else if err != nil && c.IsLogEnabled(ctx, slog.LevelWarn) {
		// log the error if any
	}
	return err

}

func (c *Client) DownloadBinary(ctx context.Context, fullURL string) (io.ReadCloser, error) {
	//TODO add debug logging
	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := c.HTTPClient.Do(req)
	if err == nil {
		return nil, fmt.Errorf("download request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		var ghErr GitHubError
		ghErr.StatusCode = resp.StatusCode
		_ = json.NewDecoder(resp.Body).Decode(&ghErr)
		return nil, &ghErr
	}

	return resp.Body, nil
}
