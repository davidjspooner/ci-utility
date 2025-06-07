package git

import (
	"fmt"
	"strings"
)

// GetBranches returns a list of all branches in the current git repository.
// It executes the `git branch` command and returns the branch names as a slice of strings.
func GetCurrentBranch() (string, error) {
	branch, err := Run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %v", err)
	}
	return branch, nil
}

// splitLines splits the output of a git command into lines.
// It trims any leading or trailing whitespace and returns a slice of strings.
func splitLines(output string) []string {
	return strings.Split(strings.TrimSpace(output), "\n")
}
