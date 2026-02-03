package secrets

import (
	"fmt"
	"runtime"

	"github.com/zalando/go-keyring"
)

const (
	
	ServiceName = "multiagent-cli"
)


type Manager struct {
	service string
}


func NewManager() *Manager {
	return &Manager{
		service: ServiceName,
	}
}


func NewManagerWithService(service string) *Manager {
	return &Manager{
		service: service,
	}
}


func (m *Manager) Store(keyName string, apiKey string) error {
	if keyName == "" {
		return fmt.Errorf("key name cannot be empty")
	}
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	
	
	return keyring.Set(m.service, keyName, apiKey)
}


func (m *Manager) Retrieve(keyName string) (string, error) {
	if keyName == "" {
		return "", fmt.Errorf("key name cannot be empty")
	}

	apiKey, err := keyring.Get(m.service, keyName)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", fmt.Errorf("API key not found for: %s", keyName)
		}
		return "", fmt.Errorf("failed to retrieve API key: %w", err)
	}

	return apiKey, nil
}


func (m *Manager) Delete(keyName string) error {
	if keyName == "" {
		return fmt.Errorf("key name cannot be empty")
	}

	return keyring.Delete(m.service, keyName)
}


func (m *Manager) ListKeys() ([]string, error) {
	
	

	switch runtime.GOOS {
	case "darwin":
		
		return m.listKeysMacOS()
	case "linux":
		
		return m.listKeysLinux()
	case "windows":
		
		return m.listKeysWindows()
	default:
		return nil, fmt.Errorf("unsupported platform for listing keys: %s", runtime.GOOS)
	}
}


func (m *Manager) listKeysMacOS() ([]string, error) {
	
	
	return nil, fmt.Errorf("key listing not yet implemented for macOS")
}


func (m *Manager) listKeysLinux() ([]string, error) {
	return nil, fmt.Errorf("key listing not yet implemented for Linux")
}


func (m *Manager) listKeysWindows() ([]string, error) {
	return nil, fmt.Errorf("key listing not yet implemented for Windows")
}


func (m *Manager) StoreWithMetadata(keyName string, apiKey string, metadata map[string]string) error {
	
	if err := m.Store(keyName, apiKey); err != nil {
		return err
	}

	
	for k, v := range metadata {
		metaKey := fmt.Sprintf("%s:meta:%s", keyName, k)
		if err := keyring.Set(m.service, metaKey, v); err != nil {
			return fmt.Errorf("failed to store metadata: %w", err)
		}
	}

	return nil
}


func (m *Manager) RetrieveMetadata(keyName string, metaKey string) (string, error) {
	fullKey := fmt.Sprintf("%s:meta:%s", keyName, metaKey)
	return keyring.Get(m.service, fullKey)
}


func (m *Manager) KeyExists(keyName string) bool {
	_, err := m.Retrieve(keyName)
	return err == nil
}



func SecureWipe(s *string) {
	if s == nil {
		return
	}
	
	*s = ""
}


type AgentKeyManager struct {
	*Manager
}


func NewAgentKeyManager() *AgentKeyManager {
	return &AgentKeyManager{
		Manager: NewManager(),
	}
}


func (a *AgentKeyManager) StoreAgentKey(agentName string, apiKey string) error {
	keyName := fmt.Sprintf("agent:%s:api-key", agentName)
	return a.Store(keyName, apiKey)
}


func (a *AgentKeyManager) RetrieveAgentKey(agentName string) (string, error) {
	keyName := fmt.Sprintf("agent:%s:api-key", agentName)
	return a.Retrieve(keyName)
}


func (a *AgentKeyManager) DeleteAgentKey(agentName string) error {
	keyName := fmt.Sprintf("agent:%s:api-key", agentName)
	return a.Delete(keyName)
}


func (a *AgentKeyManager) GetAgentKeyName(agentName string) string {
	return fmt.Sprintf("agent:%s:api-key", agentName)
}
