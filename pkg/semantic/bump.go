package semantic

import "strings"

// Bump represents a semantic version bump level and its associated commit message hints.
type Bump struct {
	Level string
	Hints []string
}

// BumpArray is a slice of Bump objects.
type BumpArray []Bump

// Bumps defines the available semantic version bump levels and their hints.
var Bumps = BumpArray{
	{Level: "major", Hints: []string{"BREAKING CHANGE", "breaking:"}},
	{Level: "minor", Hints: []string{"feat:"}},
	{Level: "patch", Hints: []string{"fix:", "chore:", "docs:", "style:", "refactor:", "perf:", "test:"}},
}

// GetVersionBump determines the version bump level based on commit messages.
// It returns the highest-priority bump found, or "patch" if none match.
func (bumps BumpArray) GetVersionBump(commits []string) (string, error) {
	found := map[string]bool{}
	for _, msg := range commits {
		for _, bump := range bumps {
			for _, hint := range bump.Hints {
				if strings.Contains(msg, hint) {
					found[bump.Level] = true
				}
			}
		}
	}

	// Check for bump levels in priority order.
	for _, bump := range bumps {
		if found[bump.Level] {
			return bump.Level, nil
		}
	}

	// Default to patch if no hints are found.
	return "patch", nil
}
