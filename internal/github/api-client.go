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

// Client is a GitHub API client for interacting with the GitHub REST API.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string // eg. "https://api.github.com/"
	Token      string
	Owner      string
	Repo       string
}

// Do sends an HTTP request to the GitHub API and decodes the response.
// It handles authentication, headers, and error responses.
func (c *Client) Do(ctx context.Context, method, fullURL string, body io.Reader, headers http.Header, response interface{}) error {
	// Replace placeholders in the URL with owner and repo.
	fullURL = strings.ReplaceAll(fullURL, "{owner}", url.PathEscape(c.Owner))
	fullURL = strings.ReplaceAll(fullURL, "{repo}", url.PathEscape(c.Repo))
	if !strings.HasPrefix(fullURL, "http") {
		fullURL = path.Join(c.BaseURL, fullURL)
	}

	// Create the HTTP request.
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication and accept headers.
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

	// Send the HTTP request.
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-2xx responses as errors.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var ghErr Error
		ghErr.StatusCode = resp.StatusCode
		_ = json.NewDecoder(resp.Body).Decode(&ghErr)
		return &ghErr
	}

	// Decode the response body if a response object is provided.
	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// DoJSON sends a JSON-encoded HTTP request to the GitHub API and decodes the JSON response.
func (c *Client) DoJSON(ctx context.Context, method, path string, request interface{}, response interface{}) error {
	var bodyReader io.Reader
	var b []byte
	var err error
	// Marshal the request body if provided.
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
	// Send the HTTP request.
	err = c.Do(ctx, method, fullURL, bodyReader, headers, response)
	if err != nil {
		// Log a warning if the request fails.
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
	// Log debug information for successful requests.
	slog.DebugContext(ctx, "GitHub API request",
		"method", method,
		"path", path,
		"request_body", string(b),
		"response", response,
		"full_url", fullURL,
	)
	return nil
}

// GetJSON sends a GET request and decodes the JSON response.
func (c *Client) GetJSON(ctx context.Context, path string, response interface{}) error {
	return c.DoJSON(ctx, http.MethodGet, path, nil, response)
}

// PostJSON sends a POST request with a JSON body and decodes the JSON response.
func (c *Client) PostJSON(ctx context.Context, path string, request interface{}, response interface{}) error {
	return c.DoJSON(ctx, http.MethodPost, path, request, response)
}

// PutJSON sends a PUT request with a JSON body and decodes the JSON response.
func (c *Client) PutJSON(ctx context.Context, path string, request interface{}, response interface{}) error {
	return c.DoJSON(ctx, http.MethodPut, path, request, response)
}

// DeleteJSON sends a DELETE request and decodes the JSON response.
func (c *Client) DeleteJSON(ctx context.Context, path string, response interface{}) error {
	return c.DoJSON(ctx, http.MethodDelete, path, nil, response)
}

// UploadMeta contains metadata for uploading a release asset to GitHub.
type UploadMeta struct {
	Name        string
	Label       string
	ContentType string // Optional, if provided will be used instead of detecting from file extension
	UploadURL   string // Optional, if provided will be used instead of constructing the URL
}

// UploadBinaryFile uploads a file as a release asset to GitHub.
func (c *Client) UploadBinaryFile(ctx context.Context, releaseID int64, fileName string, response interface{}) error {
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", fileName, err)
	}
	defer file.Close()
	// Seek to the end to get the file length.
	length, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to seek file %s: %w", fileName, err)
	}
	// Reset the file pointer to the start.
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to reset file %s: %w", fileName, err)
	}
	meta := UploadMeta{
		Name: filepath.Base(fileName),
	}
	if meta.Name == "" {
		meta.Name = path.Base(fileName)
	}
	// Upload the file stream.
	return c.UploadBinaryStream(ctx, releaseID, meta, file, length, response)
}

// UploadBinaryStream uploads a stream as a release asset to GitHub.
func (c *Client) UploadBinaryStream(ctx context.Context, releaseID int64, meta UploadMeta, data io.Reader, length int64, response interface{}) error {
	base := meta.UploadURL
	if base == "" {
		base = fmt.Sprintf("https://uploads.github.com/repos/%s/%s/releases/%d/assets", url.PathEscape(c.Owner), url.PathEscape(c.Repo), releaseID)
	}

	u, err := url.Parse(base)
	if err != nil {
		return fmt.Errorf("failed to build upload URL: %w", err)
	}

	// Set query parameters for asset name and label.
	q := u.Query()
	q.Set("name", meta.Name)
	if meta.Label != "" {
		q.Set("label", meta.Label)
	}
	u.RawQuery = q.Encode()

	// Detect content type if not provided.
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

	// Perform the upload request.
	err = c.Do(ctx, http.MethodPost, u.String(), data, headers, response)
	if err != nil {
		// Log a warning if upload fails.
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
	// Log debug information for successful upload.
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

// DownloadBinary downloads a binary file from the given URL using the GitHub API.
func (c *Client) DownloadBinary(ctx context.Context, fullURL string) (io.ReadCloser, error) {
	// DownloadBinary fetches a binary file from the provided URL using the GitHub API.
	// TODO add debug logging
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

	// Check for non-OK status code.
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		var ghErr Error
		ghErr.StatusCode = resp.StatusCode
		_ = json.NewDecoder(resp.Body).Decode(&ghErr)
		return nil, &ghErr
	}

	// Return the response body for reading.
	return resp.Body, nil
}
