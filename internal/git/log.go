package git

import (
	"fmt"
	"os/exec"
)

func GetCommitLog(branch string, limit int) (string, error) {
	cmd := exec.Command("git", "log", "--oneline", "-n", fmt.Sprintf("%d", limit), branch)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
