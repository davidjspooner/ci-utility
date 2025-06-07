package github

import (
	"path/filepath"
)

// globFiles expands a list of glob patterns into a slice of matching file paths.
func globFiles(patterns []string) ([]string, error) {
	var files []string
	for _, pattern := range patterns {
		// Use filepath.Glob to find matches for each pattern.
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		files = append(files, matches...)
	}
	// Return the collected files.
	return files, nil
}
