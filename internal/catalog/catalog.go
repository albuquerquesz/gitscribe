package catalog

import (
	"github.com/albuquerquesz/gitscribe/internal/config"
)

type ModelInfo struct {
	ID           string
	Name         string
	Provider     config.AgentProvider
	Description  string
	AuthType     AuthType
}

type AuthType string

const (
	AuthTypeOAuth   AuthType = "oauth"
	AuthTypeAPIKey  AuthType = "apikey"
)

var DefaultModels = []ModelInfo{
	{
		ID:          "claude-3-5-sonnet-20241022",
		Name:        "Claude 3.5 Sonnet",
		Provider:    config.ProviderClaude,
		Description: "Most powerful model from Anthropic",
		AuthType:    AuthTypeOAuth,
	},
	{
		ID:          "gpt-4o",
		Name:        "GPT-4o",
		Provider:    config.ProviderOpenAI,
		Description: "Omni model from OpenAI",
		AuthType:    AuthTypeOAuth,
	},
	{
		ID:          "opencode-zen",
		Name:        "OpenCode Zen",
		Provider:    config.ProviderOpenRouter,
		Description: "Optimized for coding tasks",
		AuthType:    AuthTypeAPIKey,
	},
	{
		ID:          "llama-3.3-70b-versatile",
		Name:        "Llama 3.3 70B",
		Provider:    config.ProviderGroq,
		Description: "Ultra-fast inference via Groq",
		AuthType:    AuthTypeAPIKey,
	},
}
