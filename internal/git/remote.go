package git

import (
	"os/exec"
	"strings"
)

func GetRemoteURL() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func DetectProvider(remoteURL string) string {
	lower := strings.ToLower(remoteURL)
	if strings.Contains(lower, "github.com") {
		return "github"
	}
	if strings.Contains(lower, "gitlab.com") || strings.Contains(lower, "gitlab") {
		return "gitlab"
	}
	return ""
}

func IsCLIInstalled(cli string) bool {
	cmd := exec.Command("which", cli)
	err := cmd.Run()
	return err == nil
}
