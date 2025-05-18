package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %v failed: %v", args, err)
	}
	return strings.TrimSpace(out.String()), nil
}
