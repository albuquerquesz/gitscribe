package secrets

import (
	"fmt"
	"runtime"

	"github.com/zalando/go-keyring"
)

const (
	// ServiceName is the base service name for keyring entries
	ServiceName = "multiagent-cli"
)

// Manager handles secure storage of API keys
type Manager struct {
	service string
}

// NewManager creates a new secrets manager
func NewManager() *Manager {
	return &Manager{
		service: ServiceName,
	}
}

// NewManagerWithService creates a manager with custom service name
func NewManagerWithService(service string) *Manager {
	return &Manager{
		service: service,
	}
}

// Store saves an API key to the OS keyring
func (m *Manager) Store(keyName string, apiKey string) error {
	if keyName == "" {
		return fmt.Errorf("key name cannot be empty")
	}
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// The "user" field can be used to distinguish between different keys
	// We use a fixed user "api-keys" and differentiate by keyName
	return keyring.Set(m.service, keyName, apiKey)
}

// Retrieve gets an API key from the OS keyring
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

// Delete removes an API key from the OS keyring
func (m *Manager) Delete(keyName string) error {
	if keyName == "" {
		return fmt.Errorf("key name cannot be empty")
	}

	return keyring.Delete(m.service, keyName)
}

// ListKeys returns all stored key names (note: not all keyring backends support this)
func (m *Manager) ListKeys() ([]string, error) {
	// Note: The go-keyring library doesn't provide a direct way to list keys
	// This is a platform-dependent operation

	switch runtime.GOOS {
	case "darwin":
		// macOS - could use `security` command
		return m.listKeysMacOS()
	case "linux":
		// Linux - depends on implementation (secret-service, kwallet, etc.)
		return m.listKeysLinux()
	case "windows":
		// Windows - uses wincred
		return m.listKeysWindows()
	default:
		return nil, fmt.Errorf("unsupported platform for listing keys: %s", runtime.GOOS)
	}
}

// listKeysMacOS attempts to list keys on macOS
func (m *Manager) listKeysMacOS() ([]string, error) {
	// On macOS, we could potentially use the `security` command
	// but for simplicity, we'll maintain our own registry
	return nil, fmt.Errorf("key listing not yet implemented for macOS")
}

// listKeysLinux attempts to list keys on Linux
func (m *Manager) listKeysLinux() ([]string, error) {
	return nil, fmt.Errorf("key listing not yet implemented for Linux")
}

// listKeysWindows attempts to list keys on Windows
func (m *Manager) listKeysWindows() ([]string, error) {
	return nil, fmt.Errorf("key listing not yet implemented for Windows")
}

// StoreWithMetadata saves an API key with additional metadata
func (m *Manager) StoreWithMetadata(keyName string, apiKey string, metadata map[string]string) error {
	// Store the main key
	if err := m.Store(keyName, apiKey); err != nil {
		return err
	}

	// Store metadata as separate entries if needed
	for k, v := range metadata {
		metaKey := fmt.Sprintf("%s:meta:%s", keyName, k)
		if err := keyring.Set(m.service, metaKey, v); err != nil {
			return fmt.Errorf("failed to store metadata: %w", err)
		}
	}

	return nil
}

// RetrieveMetadata gets metadata for a key
func (m *Manager) RetrieveMetadata(keyName string, metaKey string) (string, error) {
	fullKey := fmt.Sprintf("%s:meta:%s", keyName, metaKey)
	return keyring.Get(m.service, fullKey)
}

// KeyExists checks if a key exists in the keyring
func (m *Manager) KeyExists(keyName string) bool {
	_, err := m.Retrieve(keyName)
	return err == nil
}

// Secure wipe helper - attempts to clear sensitive data from memory
// Note: This is best-effort due to Go's garbage collector and string immutability
func SecureWipe(s *string) {
	if s == nil {
		return
	}
	// Strings are immutable in Go, so we can only replace the reference
	*s = ""
}

// AgentKeyManager wraps the secrets manager for agent-specific operations
type AgentKeyManager struct {
	*Manager
}

// NewAgentKeyManager creates a new agent key manager
func NewAgentKeyManager() *AgentKeyManager {
	return &AgentKeyManager{
		Manager: NewManager(),
	}
}

// StoreAgentKey stores an API key for a specific agent
func (a *AgentKeyManager) StoreAgentKey(agentName string, apiKey string) error {
	keyName := fmt.Sprintf("agent:%s:api-key", agentName)
	return a.Store(keyName, apiKey)
}

// RetrieveAgentKey retrieves an API key for a specific agent
func (a *AgentKeyManager) RetrieveAgentKey(agentName string) (string, error) {
	keyName := fmt.Sprintf("agent:%s:api-key", agentName)
	return a.Retrieve(keyName)
}

// DeleteAgentKey removes an API key for a specific agent
func (a *AgentKeyManager) DeleteAgentKey(agentName string) error {
	keyName := fmt.Sprintf("agent:%s:api-key", agentName)
	return a.Delete(keyName)
}

// GetAgentKeyName returns the keyring key name for an agent
func (a *AgentKeyManager) GetAgentKeyName(agentName string) string {
	return fmt.Sprintf("agent:%s:api-key", agentName)
}
