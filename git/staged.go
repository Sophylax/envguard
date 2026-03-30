package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

var execCommand = exec.Command

// StagedFiles returns staged file paths from git diff --cached --name-only.
func StagedFiles() ([]string, error) {
	cmd := execCommand("git", "diff", "--cached", "--name-only")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("run git diff --cached --name-only: %w", err)
	}

	lines := strings.Split(stdout.String(), "\n")
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		files = append(files, line)
	}
	return files, nil
}
