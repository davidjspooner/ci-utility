package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Run executes a git command with the provided arguments and returns the output.
func Run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %v failed: %v", args, err)
	}
	// Trim any leading or trailing whitespace from the output
	return strings.TrimSpace(out.String()), nil
}
