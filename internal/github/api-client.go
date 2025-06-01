package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
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

func (c *Client) Do(ctx context.Context, method, fullURL string, body io.Reader, headers http.Header, response interface{}) error {
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
	for key, value := range headers {
		for _, v := range value {
			if key == "Content-Length" {
				req.ContentLength, _ = strconv.ParseInt(v, 10, 64)
				continue
			}
			req.Header.Add(key, v)
		}
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
	headers := http.Header{
		"Content-Type":   []string{"application/json"},
		"Content-Length": []string{fmt.Sprintf("%d", len(b))},
	}
	err = c.Do(ctx, method, fullURL, bodyReader, headers, response)
	if err != nil {
		slog.WarnContext(ctx, "GitHub API request failed",
			"method", method,
			"path", path,
			"request_body", string(b),
			"response", response,
			"full_url", fullURL,
			"error", err,
		)
		return err
	}
	slog.DebugContext(ctx, "GitHub API request",
		"method", method,
		"path", path,
		"request_body", string(b),
		"response", response,
		"full_url", fullURL,
	)
	return nil
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
	Name        string
	Label       string
	ContentType string // Optional, if provided will be used instead of detecting from file extension
	UploadURL   string // Optional, if provided will be used instead of constructing the URL
}

func (c *Client) UploadBinaryFile(ctx context.Context, releaseID int64, fileName string, response interface{}) error {
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", fileName, err)
	}
	defer file.Close()
	length, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to seek file %s: %w", fileName, err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to reset file %s: %w", fileName, err)
	}
	meta := UploadMeta{
		Name: filepath.Base(fileName),
	}
	if meta.Name == "" {
		meta.Name = path.Base(fileName)
	}
	return c.UploadBinaryStream(ctx, releaseID, meta, file, length, response)
}

func (c *Client) UploadBinaryStream(ctx context.Context, releaseID int64, meta UploadMeta, data io.Reader, length int64, response interface{}) error {
	base := meta.UploadURL
	if base == "" {
		base = fmt.Sprintf("https://uploads.github.com/repos/%s/%s/releases/%d/assets", url.PathEscape(c.Owner), url.PathEscape(c.Repo), releaseID)
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

	if meta.ContentType == "" {
		meta.ContentType = mime.TypeByExtension(filepath.Ext(meta.Name))
		if meta.ContentType == "" {
			meta.ContentType = "application/octet-stream"
		}
	}
	headers := http.Header{
		"Content-Type":   []string{meta.ContentType},
		"Content-Length": []string{fmt.Sprintf("%d", length)},
	}

	err = c.Do(ctx, http.MethodPost, u.String(), data, headers, response)
	if err != nil {
		slog.WarnContext(ctx, "Failed to upload binary",
			"release_id", releaseID,
			"name", meta.Name,
			"label", meta.Label,
			"url", u.String(),
			"content_type", meta.ContentType,
			"length", length,
			"error", err,
		)
		return err
	}
	slog.DebugContext(ctx, "Uploaded binary",
		"release_id", releaseID,
		"name", meta.Name,
		"label", meta.Label,
		"url", u.String(),
		"content_type", meta.ContentType,
		"length", length,
	)
	return nil

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
