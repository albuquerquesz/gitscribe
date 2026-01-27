package git

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func StageFiles(files []string) error {
	if len(files) == 0 {
		return nil
	}

	args := append([]string{"add"}, files...)
	
	cmd := exec.Command("git", args...)

	cmd.Stdout = io.Discard

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add failed: %s", stderr.String())
	}

	return nil
}


// GetStagedDiff returns the diff of staged changes.
func GetStagedDiff() (string, error) {
	var diffOutput bytes.Buffer
	cmd := exec.Command("git", "diff", "--staged")
	cmd.Stdout = &diffOutput
	cmd.Stderr = &diffOutput

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get git diff: %s", diffOutput.String())
	}
	return diffOutput.String(), nil
}

// Commit performs a git commit with the given message.
func Commit(message string) error {
	var output bytes.Buffer
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error while committing: %s", output.String())
	}
	return nil
}

// Push pushes the changes to the specified branch on origin.
func Push(branch string) error {
	var output bytes.Buffer
	cmd := exec.Command("git", "push", "origin", branch)
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error while pushing: %s", output.String())
	}
	return nil
}

// GetCurrentBranch returns the name of the current git branch.
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}
