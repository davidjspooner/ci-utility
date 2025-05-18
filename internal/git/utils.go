package git

import (
	"fmt"
	"strings"
)

func GetCurrentBranch() (string, error) {
	branch, err := Run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %v", err)
	}
	return branch, nil
}

func splitLines(output string) []string {
	return strings.Split(strings.TrimSpace(output), "\n")
}
