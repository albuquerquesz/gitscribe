package git

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func Stage(files []string) error {
	for _, file := range files {
		cmd := exec.Command("git", "add", file)
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to stage file %s: %v", file, err)
		}
	}
	return nil
}

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

func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}
