package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ErrNoStagedFiles is returned when there are no staged changes.
var ErrNoStagedFiles = errors.New("no staged files: run `git add` before llmcommit")

// GetStagedDiff returns the output of `git diff --cached`.
// Returns ErrNoStagedFiles if the diff is empty.
func GetStagedDiff() (string, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return "", fmt.Errorf("git not found: please install git and ensure it is on your PATH")
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "diff", "--cached")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "not a git repository") {
			return "", fmt.Errorf("not in a git repository: run `git init` or change to a git repository directory")
		}
		return "", fmt.Errorf("git diff --cached failed: %w", err)
	}

	diff := stdout.String()
	if diff == "" {
		return "", ErrNoStagedFiles
	}

	return diff, nil
}

// SplitDiffByFile splits a unified diff string into per-file sections.
// Keys are file paths (parsed from `diff --git a/... b/...` headers).
func SplitDiffByFile(diff string) map[string]string {
	result := make(map[string]string)

	lines := strings.Split(diff, "\n")
	var currentFile string
	var currentLines []string

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git ") {
			// Save the previous file's diff
			if currentFile != "" {
				result[currentFile] = strings.Join(currentLines, "\n")
			}

			// Parse the file path from the header: "diff --git a/<path> b/<path>"
			// We use the b/ path as the canonical file name
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				bPath := parts[3]
				if strings.HasPrefix(bPath, "b/") {
					currentFile = bPath[2:]
				} else {
					currentFile = bPath
				}
			} else {
				currentFile = line
			}
			currentLines = []string{line}
		} else if currentFile != "" {
			currentLines = append(currentLines, line)
		}
	}

	// Save the last file's diff
	if currentFile != "" {
		result[currentFile] = strings.Join(currentLines, "\n")
	}

	return result
}
