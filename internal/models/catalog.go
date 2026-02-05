package models

import (
	"fmt"
	"strings"
)

type ModelInfo struct {
	ID          string
	Name        string
	Provider    string
	Description string
	MaxTokens   int
	Group       string
}

type ProviderInfo struct {
	Name           string
	DisplayName    string
	Icon           string
	SupportsOAuth2 bool
	Description    string
}

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
	"hackclub": {
		Name:           "hackclub",
		DisplayName:    "Hack Club",
		Icon:           "🚩",
		SupportsOAuth2: false,
		Description:    "Hack Club AI proxy",
	},
	"opencode": {
		Name:           "opencode",
		DisplayName:    "OpenCode",
		Icon:           "💻",
		SupportsOAuth2: false,
		Description:    "OpenCode AI API",
	},
}

var ModelCatalog = map[string][]ModelInfo{
	"anthropic": {
		{
			ID:          "claude-3-5-sonnet-20241022",
			Name:        "Claude 3.5 Sonnet",
			Provider:    "anthropic",
			Description: "Best balance of intelligence and speed",
			MaxTokens:   200000,
			Group:       "Claude 3.5",
		},
		{
			ID:          "claude-3-opus-20240229",
			Name:        "Claude 3 Opus",
			Provider:    "anthropic",
			Description: "Most powerful model for complex tasks",
			MaxTokens:   200000,
			Group:       "Claude 3",
		},
		{
			ID:          "claude-3-sonnet-20240229",
			Name:        "Claude 3 Sonnet",
			Provider:    "anthropic",
			Description: "Balanced performance and cost",
			MaxTokens:   200000,
			Group:       "Claude 3",
		},
		{
			ID:          "claude-3-haiku-20240307",
			Name:        "Claude 3 Haiku",
			Provider:    "anthropic",
			Description: "Fastest, most cost-effective",
			MaxTokens:   200000,
			Group:       "Claude 3",
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
		},
		{
			ID:          "gpt-4o-mini",
			Name:        "GPT-4o Mini",
			Provider:    "openai",
			Description: "Fast, affordable for most tasks",
			MaxTokens:   128000,
			Group:       "GPT-4",
		},
		{
			ID:          "gpt-4-turbo",
			Name:        "GPT-4 Turbo",
			Provider:    "openai",
			Description: "Legacy high quality model",
			MaxTokens:   128000,
			Group:       "GPT-4",
		},
		{
			ID:          "o1",
			Name:        "o1",
			Provider:    "openai",
			Description: "Advanced reasoning model",
			MaxTokens:   200000,
			Group:       "o1 Series",
		},
		{
			ID:          "o1-mini",
			Name:        "o1 Mini",
			Provider:    "openai",
			Description: "Fast reasoning model",
			MaxTokens:   128000,
			Group:       "o1 Series",
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
		},
		{
			ID:          "openai/gpt-oss-120b",
			Name:        "OpenAI GPT OSS 120b",
			Provider:    "groq",
			Description: "Ultra-fast inference with Llama",
			MaxTokens:   32768,
			Group:       "Llama 3",
		},
		{
			ID:          "mixtral-8x7b-32768",
			Name:        "Mixtral 8x7B",
			Provider:    "groq",
			Description: "Efficient MoE architecture",
			MaxTokens:   32768,
			Group:       "Mixtral",
		},
		{
			ID:          "gemma-2-9b-it",
			Name:        "Gemma 2 9B",
			Provider:    "groq",
			Description: "Lightweight Google model",
			MaxTokens:   8192,
			Group:       "Gemma",
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
		},
		{
			ID:          "codellama",
			Name:        "CodeLlama",
			Provider:    "ollama",
			Description: "Specialized for coding tasks",
			MaxTokens:   100000,
			Group:       "Code",
		},
		{
			ID:          "mistral",
			Name:        "Mistral",
			Provider:    "ollama",
			Description: "Efficient open model",
			MaxTokens:   32768,
			Group:       "Mistral",
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
		},
		{
			ID:          "openai/gpt-4o",
			Name:        "GPT-4o",
			Provider:    "openrouter",
			Description: "Via OpenRouter",
			MaxTokens:   128000,
			Group:       "OpenAI",
		},
		{
			ID:          "meta-llama/llama-3.3-70b-instruct",
			Name:        "Llama 3.3 70B",
			Provider:    "openrouter",
			Description: "Via OpenRouter",
			MaxTokens:   131072,
			Group:       "Meta",
		},
	},
	"hackclub": {
		{
			ID:          "moonshotai/kimi-k2.5",
			Name:        "Kimi K2.5",
			Provider:    "hackclub",
			Description: "Via hackclub",
			Group:       "hackclub",
		},
		{
			ID:          "qwen/qwen-2.5-72b-instruct",
			Name:        "Qwen 2.5 72B",
			Provider:    "hackclub",
			Description: "Via hackclub",
			Group:       "hackclub",
		},
	},
	"opencode": {
		{
			ID:          "opencode-default",
			Name:        "OpenCode Default",
			Provider:    "opencode",
			Description: "Default OpenCode model",
			MaxTokens:   4096,
			Group:       "OpenCode",
		},
	},
}

func GetModelsForProvider(provider string) []ModelInfo {
	if models, ok := ModelCatalog[provider]; ok {
		return models
	}
	return []ModelInfo{}
}

func GetModelByID(provider, modelID string) (ModelInfo, error) {
	models := GetModelsForProvider(provider)
	for _, m := range models {
		if m.ID == modelID {
			return m, nil
		}
	}
	return ModelInfo{}, fmt.Errorf("model %s not found for provider %s", modelID, provider)
}

func GetAllModels() []ModelInfo {
	var all []ModelInfo
	for _, models := range ModelCatalog {
		all = append(all, models...)
	}
	return all
}

func GetProviderKeys() []string {
	keys := make([]string, 0, len(Providers))
	for k := range Providers {
		keys = append(keys, k)
	}
	return keys
}

func SupportsOAuth2(provider string) bool {
	if p, ok := Providers[provider]; ok {
		return p.SupportsOAuth2
	}
	return false
}

func GenerateProfileName(provider, modelID string) string {
	cleanModel := strings.ReplaceAll(modelID, "-", "_")
	cleanModel = strings.ReplaceAll(cleanModel, ".", "_")
	cleanModel = strings.ReplaceAll(cleanModel, "/", "_")

	if idx := strings.LastIndex(cleanModel, "_20"); idx != -1 {
		cleanModel = cleanModel[:idx]
	}

	return fmt.Sprintf("%s_%s", provider, cleanModel)
}
