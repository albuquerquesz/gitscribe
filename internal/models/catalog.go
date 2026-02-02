package models

import (
	"fmt"
	"strings"
)

// ModelInfo contains metadata about an AI model
type ModelInfo struct {
	ID          string
	Name        string
	Provider    string
	Description string
	MaxTokens   int
	Group       string
	Tags        []string
}

// ProviderInfo contains information about a provider
type ProviderInfo struct {
	Name           string
	DisplayName    string
	Icon           string
	SupportsOAuth2 bool
	Description    string
}

// Available providers
var Providers = map[string]ProviderInfo{
	"anthropic": {
		Name:           "anthropic",
		DisplayName:    "Anthropic",
		Icon:           "🧠",
		SupportsOAuth2: true,
		Description:    "Claude models with excellent reasoning",
	},
	"openai": {
		Name:           "openai",
		DisplayName:    "OpenAI",
		Icon:           "🤖",
		SupportsOAuth2: false,
		Description:    "GPT models, fast and capable",
	},
	"groq": {
		Name:           "groq",
		DisplayName:    "Groq",
		Icon:           "⚡",
		SupportsOAuth2: false,
		Description:    "Ultra-fast Llama inference",
	},
	"gemini": {
		Name:           "gemini",
		DisplayName:    "Gemini",
		Icon:           "🔮",
		SupportsOAuth2: false,
		Description:    "Google multimodal models",
	},
	"ollama": {
		Name:           "ollama",
		DisplayName:    "Ollama",
		Icon:           "🦙",
		SupportsOAuth2: false,
		Description:    "Local self-hosted models",
	},
	"openrouter": {
		Name:           "openrouter",
		DisplayName:    "OpenRouter",
		Icon:           "🌐",
		SupportsOAuth2: false,
		Description:    "Multi-provider access",
	},
}

// ModelCatalog contains all available models grouped by provider
var ModelCatalog = map[string][]ModelInfo{
	"anthropic": {
		{
			ID:          "claude-3-5-sonnet-20241022",
			Name:        "Claude 3.5 Sonnet",
			Provider:    "anthropic",
			Description: "Best balance of intelligence and speed",
			MaxTokens:   200000,
			Group:       "Claude 3.5",
			Tags:        []string{"fast", "smart"},
		},
		{
			ID:          "claude-3-opus-20240229",
			Name:        "Claude 3 Opus",
			Provider:    "anthropic",
			Description: "Most powerful model for complex tasks",
			MaxTokens:   200000,
			Group:       "Claude 3",
			Tags:        []string{"powerful", "slow"},
		},
		{
			ID:          "claude-3-sonnet-20240229",
			Name:        "Claude 3 Sonnet",
			Provider:    "anthropic",
			Description: "Balanced performance and cost",
			MaxTokens:   200000,
			Group:       "Claude 3",
			Tags:        []string{"balanced"},
		},
		{
			ID:          "claude-3-haiku-20240307",
			Name:        "Claude 3 Haiku",
			Provider:    "anthropic",
			Description: "Fastest, most cost-effective",
			MaxTokens:   200000,
			Group:       "Claude 3",
			Tags:        []string{"fast", "cheap"},
		},
	},
	"openai": {
		{
			ID:          "gpt-4o",
			Name:        "GPT-4o",
			Provider:    "openai",
			Description: "Most capable multimodal model",
			MaxTokens:   128000,
			Group:       "GPT-4",
			Tags:        []string{"smart", "multimodal"},
		},
		{
			ID:          "gpt-4o-mini",
			Name:        "GPT-4o Mini",
			Provider:    "openai",
			Description: "Fast, affordable for most tasks",
			MaxTokens:   128000,
			Group:       "GPT-4",
			Tags:        []string{"fast", "cheap"},
		},
		{
			ID:          "gpt-4-turbo",
			Name:        "GPT-4 Turbo",
			Provider:    "openai",
			Description: "Legacy high quality model",
			MaxTokens:   128000,
			Group:       "GPT-4",
			Tags:        []string{"legacy"},
		},
		{
			ID:          "o1",
			Name:        "o1",
			Provider:    "openai",
			Description: "Advanced reasoning model",
			MaxTokens:   200000,
			Group:       "o1 Series",
			Tags:        []string{"reasoning", "slow"},
		},
		{
			ID:          "o1-mini",
			Name:        "o1 Mini",
			Provider:    "openai",
			Description: "Fast reasoning model",
			MaxTokens:   128000,
			Group:       "o1 Series",
			Tags:        []string{"reasoning", "fast"},
		},
	},
	"groq": {
		{
			ID:          "llama-3.3-70b-versatile",
			Name:        "Llama 3.3 70B",
			Provider:    "groq",
			Description: "Ultra-fast inference with Llama",
			MaxTokens:   32768,
			Group:       "Llama 3",
			Tags:        []string{"fast", "local"},
		},
		{
			ID:          "mixtral-8x7b-32768",
			Name:        "Mixtral 8x7B",
			Provider:    "groq",
			Description: "Efficient MoE architecture",
			MaxTokens:   32768,
			Group:       "Mixtral",
			Tags:        []string{"efficient"},
		},
		{
			ID:          "gemma-2-9b-it",
			Name:        "Gemma 2 9B",
			Provider:    "groq",
			Description: "Lightweight Google model",
			MaxTokens:   8192,
			Group:       "Gemma",
			Tags:        []string{"lightweight"},
		},
	},
	"gemini": {
		{
			ID:          "gemini-1.5-pro",
			Name:        "Gemini 1.5 Pro",
			Provider:    "gemini",
			Description: "Google's most capable model",
			MaxTokens:   2000000,
			Group:       "Gemini 1.5",
			Tags:        []string{"smart", "long-context"},
		},
		{
			ID:          "gemini-1.5-flash",
			Name:        "Gemini 1.5 Flash",
			Provider:    "gemini",
			Description: "Fast and versatile",
			MaxTokens:   1000000,
			Group:       "Gemini 1.5",
			Tags:        []string{"fast"},
		},
	},
	"ollama": {
		{
			ID:          "llama3.2",
			Name:        "Llama 3.2",
			Provider:    "ollama",
			Description: "Meta's latest local model",
			MaxTokens:   128000,
			Group:       "Llama",
			Tags:        []string{"local", "free"},
		},
		{
			ID:          "codellama",
			Name:        "CodeLlama",
			Provider:    "ollama",
			Description: "Specialized for coding tasks",
			MaxTokens:   100000,
			Group:       "Code",
			Tags:        []string{"local", "code"},
		},
		{
			ID:          "mistral",
			Name:        "Mistral",
			Provider:    "ollama",
			Description: "Efficient open model",
			MaxTokens:   32768,
			Group:       "Mistral",
			Tags:        []string{"local", "efficient"},
		},
	},
	"openrouter": {
		{
			ID:          "anthropic/claude-3.5-sonnet",
			Name:        "Claude 3.5 Sonnet",
			Provider:    "openrouter",
			Description: "Via OpenRouter",
			MaxTokens:   200000,
			Group:       "Anthropic",
			Tags:        []string{"smart"},
		},
		{
			ID:          "openai/gpt-4o",
			Name:        "GPT-4o",
			Provider:    "openrouter",
			Description: "Via OpenRouter",
			MaxTokens:   128000,
			Group:       "OpenAI",
			Tags:        []string{"smart"},
		},
		{
			ID:          "meta-llama/llama-3.3-70b-instruct",
			Name:        "Llama 3.3 70B",
			Provider:    "openrouter",
			Description: "Via OpenRouter",
			MaxTokens:   131072,
			Group:       "Meta",
			Tags:        []string{"fast"},
		},
	},
}

// GetModelsForProvider returns all models for a given provider
func GetModelsForProvider(provider string) []ModelInfo {
	if models, ok := ModelCatalog[provider]; ok {
		return models
	}
	return []ModelInfo{}
}

// GetModelByID returns a specific model by its ID
func GetModelByID(provider, modelID string) (ModelInfo, error) {
	models := GetModelsForProvider(provider)
	for _, m := range models {
		if m.ID == modelID {
			return m, nil
		}
	}
	return ModelInfo{}, fmt.Errorf("model %s not found for provider %s", modelID, provider)
}

// GetAllModels returns all models from all providers
func GetAllModels() []ModelInfo {
	var all []ModelInfo
	for _, models := range ModelCatalog {
		all = append(all, models...)
	}
	return all
}

// GetProviderKeys returns all provider names
func GetProviderKeys() []string {
	keys := make([]string, 0, len(Providers))
	for k := range Providers {
		keys = append(keys, k)
	}
	return keys
}

// SupportsOAuth2 checks if a provider supports OAuth2 authentication
func SupportsOAuth2(provider string) bool {
	if p, ok := Providers[provider]; ok {
		return p.SupportsOAuth2
	}
	return false
}

// GenerateProfileName creates a profile name from provider and model
func GenerateProfileName(provider, modelID string) string {
	// Remove version dates and special chars
	cleanModel := strings.ReplaceAll(modelID, "-", "_")
	cleanModel = strings.ReplaceAll(cleanModel, ".", "_")
	cleanModel = strings.ReplaceAll(cleanModel, "/", "_")

	// Remove date patterns like 20241022
	if idx := strings.LastIndex(cleanModel, "_20"); idx != -1 {
		cleanModel = cleanModel[:idx]
	}

	return fmt.Sprintf("%s_%s", provider, cleanModel)
}
