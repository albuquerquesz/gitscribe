package ai

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/albuquerquesz/gitscribe/internal/config"
)

func BuildPromptWithContext(baseDiff, projectPath string) string {
	if projectPath == "" {
		return baseDiff
	}

	cm, err := config.LoadContexts()
	if err != nil {
		return baseDiff
	}

	contexts := cm.GetContextsForPrompt(projectPath)
	if contexts == "" {
		return baseDiff
	}

	return fmt.Sprintf(`Contextos adicionais do projeto:
%s

Analise o diff abaixo considerando os contextos acima:

%s`, contexts, baseDiff)
}

func GetCurrentProjectPath() string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
