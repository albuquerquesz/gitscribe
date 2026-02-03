package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	ConfigDirName  = ".multiagent"
	ConfigFileName = "config.yaml"
)


type AgentProvider string

const (
	ProviderOpenAI     AgentProvider = "openai"
	ProviderGroq       AgentProvider = "groq"
	ProviderClaude     AgentProvider = "claude"
	ProviderGemini     AgentProvider = "gemini"
	ProviderOllama     AgentProvider = "ollama"
	ProviderOpenRouter AgentProvider = "openrouter"
)


type AgentProfile struct {
	Name         string        `yaml:"name" json:"name"`
	Provider     AgentProvider `yaml:"provider" json:"provider"`
	Model        string        `yaml:"model" json:"model"`
	BaseURL      string        `yaml:"base_url,omitempty" json:"base_url,omitempty"`
	Temperature  float32       `yaml:"temperature" json:"temperature"`
	MaxTokens    int           `yaml:"max_tokens" json:"max_tokens"`
	Timeout      int           `yaml:"timeout_seconds" json:"timeout_seconds"`
	Enabled      bool          `yaml:"enabled" json:"enabled"`
	Priority     int           `yaml:"priority" json:"priority"` 
	SystemPrompt string        `yaml:"system_prompt,omitempty" json:"system_prompt,omitempty"`
	
	KeyringKey string `yaml:"keyring_key" json:"keyring_key"` 
}


type RoutingRule struct {
	Name         string   `yaml:"name" json:"name"`
	AgentProfile string   `yaml:"agent_profile" json:"agent_profile"`
	Conditions   []string `yaml:"conditions" json:"conditions"`
	Priority     int      `yaml:"priority" json:"priority"`
}


type GlobalConfig struct {
	DefaultAgent   string            `yaml:"default_agent" json:"default_agent"`
	AutoSelect     bool              `yaml:"auto_select" json:"auto_select"`
	RequestTimeout int               `yaml:"request_timeout_seconds" json:"request_timeout_seconds"`
	MaxRetries     int               `yaml:"max_retries" json:"max_retries"`
	LogLevel       string            `yaml:"log_level" json:"log_level"`
	CustomHeaders  map[string]string `yaml:"custom_headers,omitempty" json:"custom_headers,omitempty"`
}


type Config struct {
	Version string         `yaml:"version" json:"version"`
	Global  GlobalConfig   `yaml:"global" json:"global"`
	Agents  []AgentProfile `yaml:"agents" json:"agents"`
	Routing []RoutingRule  `yaml:"routing" json:"routing"`
}


func DefaultConfig() *Config {
	return &Config{
		Version: "1.0",
		Global: GlobalConfig{
			DefaultAgent:   "groq-default",
			AutoSelect:     true,
			RequestTimeout: 30,
			MaxRetries:     3,
			LogLevel:       "info",
		},
		Agents: []AgentProfile{
			{
				Name:         "groq-default",
				Provider:     ProviderGroq,
				Model:        "llama-3.3-70b-versatile",
				Temperature:  0.7,
				MaxTokens:    2048,
				Timeout:      30,
				Enabled:      true,
				Priority:     1,
				SystemPrompt: "You are a helpful assistant.",
				KeyringKey:   "groq-default-api-key",
			},
			{
				Name:         "openai-gpt4",
				Provider:     ProviderOpenAI,
				Model:        "gpt-4",
				Temperature:  0.7,
				MaxTokens:    4096,
				Timeout:      60,
				Enabled:      false,
				Priority:     2,
				SystemPrompt: "You are a helpful assistant.",
				KeyringKey:   "openai-gpt4-api-key",
			},
			{
				Name:         "claude-sonnet",
				Provider:     ProviderClaude,
				Model:        "claude-3-5-sonnet-20241022",
				Temperature:  0.7,
				MaxTokens:    8192,
				Timeout:      60,
				Enabled:      false,
				Priority:     3,
				SystemPrompt: "You are a helpful assistant with strong reasoning capabilities.",
				KeyringKey:   "claude-sonnet-api-key",
			},
		},
		Routing: []RoutingRule{
			{
				Name:         "quick-tasks",
				AgentProfile: "groq-default",
				Conditions:   []string{"token_count < 1000", "complexity = low"},
				Priority:     1,
			},
			{
				Name:         "complex-reasoning",
				AgentProfile: "claude-sonnet",
				Conditions:   []string{"complexity = high", "reasoning = required"},
				Priority:     2,
			},
		},
	}
}


func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ConfigDirName)
	return filepath.Join(configDir, ConfigFileName), nil
}


func EnsureConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ConfigDirName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return configDir, nil
}


func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}


func (c *Config) Save() error {
	_, err := EnsureConfigDir()
	if err != nil {
		return err
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}


func (c *Config) GetAgentByName(name string) (*AgentProfile, error) {
	for i := range c.Agents {
		if c.Agents[i].Name == name {
			return &c.Agents[i], nil
		}
	}
	return nil, fmt.Errorf("agent profile not found: %s", name)
}


func (c *Config) GetDefaultAgent() (*AgentProfile, error) {
	if c.Global.DefaultAgent == "" {
		
		for i := range c.Agents {
			if c.Agents[i].Enabled {
				return &c.Agents[i], nil
			}
		}
		return nil, fmt.Errorf("no enabled agents found")
	}

	return c.GetAgentByName(c.Global.DefaultAgent)
}


func (c *Config) ListEnabledAgents() []AgentProfile {
	var enabled []AgentProfile
	for _, agent := range c.Agents {
		if agent.Enabled {
			enabled = append(enabled, agent)
		}
	}
	return enabled
}


func (c *Config) AddAgent(agent AgentProfile) error {
	
	if _, err := c.GetAgentByName(agent.Name); err == nil {
		return fmt.Errorf("agent profile already exists: %s", agent.Name)
	}

	c.Agents = append(c.Agents, agent)
	return nil
}


func (c *Config) RemoveAgent(name string) error {
	for i, agent := range c.Agents {
		if agent.Name == name {
			c.Agents = append(c.Agents[:i], c.Agents[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("agent profile not found: %s", name)
}


func (c *Config) SetDefaultAgent(name string) error {
	if _, err := c.GetAgentByName(name); err != nil {
		return err
	}
	c.Global.DefaultAgent = name
	return nil
}
