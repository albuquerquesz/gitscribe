package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	MaxContextsPerPath = 3
	contextsFileName   = "contexts.json"
)

type ContextEntry struct {
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type ContextManager struct {
	Contexts map[string][]ContextEntry `json:"contexts"`
}

func getContextsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "gitscribe", contextsFileName)
}

func LoadContexts() (*ContextManager, error) {
	path := getContextsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ContextManager{Contexts: make(map[string][]ContextEntry)}, nil
		}
		return nil, err
	}

	var cm ContextManager
	if err := json.Unmarshal(data, &cm); err != nil {
		return nil, err
	}

	if cm.Contexts == nil {
		cm.Contexts = make(map[string][]ContextEntry)
	}

	return &cm, nil
}

func (cm *ContextManager) Save() error {
	path := getContextsPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cm, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (cm *ContextManager) AddContext(projectPath, text string) error {
	contexts := cm.Contexts[projectPath]
	if len(contexts) >= MaxContextsPerPath {
		return fmt.Errorf("limite de %d contextos atingido para este projeto", MaxContextsPerPath)
	}

	cm.Contexts[projectPath] = append(contexts, ContextEntry{
		Text:      text,
		CreatedAt: time.Now(),
	})

	return cm.Save()
}

func (cm *ContextManager) RemoveContext(projectPath string, index int) error {
	contexts, exists := cm.Contexts[projectPath]
	if !exists || index < 0 || index >= len(contexts) {
		return fmt.Errorf("índice inválido")
	}

	cm.Contexts[projectPath] = append(contexts[:index], contexts[index+1:]...)

	if len(cm.Contexts[projectPath]) == 0 {
		delete(cm.Contexts, projectPath)
	}

	return cm.Save()
}

func (cm *ContextManager) ListContexts(projectPath string) []ContextEntry {
	return cm.Contexts[projectPath]
}

func (cm *ContextManager) GetContextsForPrompt(projectPath string) string {
	contexts := cm.Contexts[projectPath]
	if len(contexts) == 0 {
		return ""
	}

	var result string
	for i, ctx := range contexts {
		if i > 0 {
			result += "\n"
		}
		result += fmt.Sprintf("- %s", ctx.Text)
	}
	return result
}

func (cm *ContextManager) GetAllPaths() []string {
	paths := make([]string, 0, len(cm.Contexts))
	for path := range cm.Contexts {
		paths = append(paths, path)
	}
	return paths
}

func (cm *ContextManager) RemovePath(projectPath string) error {
	delete(cm.Contexts, projectPath)
	return cm.Save()
}

func (cm *ContextManager) CleanNonExistentPaths() error {
	for path := range cm.Contexts {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			delete(cm.Contexts, path)
		}
	}
	return cm.Save()
}
