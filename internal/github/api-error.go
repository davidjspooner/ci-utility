package github

import (
	"fmt"
)

// Error represents an error returned by the GitHub API.
type Error struct {
	StatusCode int
	Message    string                   `json:"message"`
	Errors     []map[string]interface{} `json:"errors,omitempty"`
	DocsURL    string                   `json:"documentation_url,omitempty"`
}

// Error implements the error interface for GitHubError.
func (e *Error) Error() string {
	return fmt.Sprintf("GitHub API error (%d): %s", e.StatusCode, e.Message)
}
