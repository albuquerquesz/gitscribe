package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// anthropicProvider implements ModelProvider for Anthropic
type anthropicProvider struct {
	config ProviderConfig
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider() ModelProvider {
	return &anthropicProvider{
		config: ProviderConfigs["anthropic"],
	}
}

func (p *anthropicProvider) Name() string {
	return "anthropic"
}

func (p *anthropicProvider) Config() ProviderConfig {
	return p.config
}

func (p *anthropicProvider) SupportsDynamicFetch() bool {
	return false // Anthropic doesn't expose a public models endpoint
}

func (p *anthropicProvider) FetchModels(ctx context.Context, apiKey string) ([]Model, error) {
	// Anthropic doesn't support dynamic fetching
	return nil, fmt.Errorf("anthropic does not support dynamic model fetching")
}

func (p *anthropicProvider) ValidateAPIKey(ctx context.Context, apiKey string) error {
	// Simple validation - make a request to list models (which requires auth)
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "GET", p.config.BaseURL+"/models", nil)
	if err != nil {
		return err
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate API key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid API key")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API key validation failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (p *anthropicProvider) GetDefaultModels() []Model {
	return GetStaticModels("anthropic")
}

// openAIProvider implements ModelProvider for OpenAI
type openAIProvider struct {
	config ProviderConfig
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider() ModelProvider {
	return &openAIProvider{
		config: ProviderConfigs["openai"],
	}
}

func (p *openAIProvider) Name() string {
	return "openai"
}

func (p *openAIProvider) Config() ProviderConfig {
	return p.config
}

func (p *openAIProvider) SupportsDynamicFetch() bool {
	return true
}

func (p *openAIProvider) FetchModels(ctx context.Context, apiKey string) ([]Model, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "GET", p.config.BaseURL+"/models", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch models, status: %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode models response: %w", err)
	}

	models := make([]Model, 0, len(result.Data))
	for _, m := range result.Data {
		// Skip non-chat models
		if !isChatModel(m.ID) {
			continue
		}

		model := Model{
			ID:        m.ID,
			Provider:  "openai",
			Name:      m.ID,
			Status:    ModelStatusAvailable,
			CreatedAt: m.Created,
		}

		// Try to match with static data for richer info
		if static, err := findStaticModel("openai", m.ID); err == nil {
			model.Name = static.Name
			model.Description = static.Description
			model.Capabilities = static.Capabilities
			model.PricingTier = static.PricingTier
			model.ContextWindow = static.ContextWindow
			model.MaxTokens = static.MaxTokens
			model.InputPrice = static.InputPrice
			model.OutputPrice = static.OutputPrice
			model.SupportsVision = static.SupportsVision
			model.SupportsToolUse = static.SupportsToolUse
		}

		models = append(models, model)
	}

	return models, nil
}

func (p *openAIProvider) ValidateAPIKey(ctx context.Context, apiKey string) error {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "GET", p.config.BaseURL+"/models", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate API key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid API key")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API key validation failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (p *openAIProvider) GetDefaultModels() []Model {
	return GetStaticModels("openai")
}

// isChatModel filters for chat-capable models
func isChatModel(id string) bool {
	chatPrefixes := []string{
		"gpt-", "o1-", "text-embedding-",
	}
	for _, prefix := range chatPrefixes {
		if len(id) >= len(prefix) && id[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// groqProvider implements ModelProvider for Groq
type groqProvider struct {
	config ProviderConfig
}

// NewGroqProvider creates a new Groq provider
func NewGroqProvider() ModelProvider {
	return &groqProvider{
		config: ProviderConfigs["groq"],
	}
}

func (p *groqProvider) Name() string {
	return "groq"
}

func (p *groqProvider) Config() ProviderConfig {
	return p.config
}

func (p *groqProvider) SupportsDynamicFetch() bool {
	return true
}

func (p *groqProvider) FetchModels(ctx context.Context, apiKey string) ([]Model, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "GET", p.config.BaseURL+"/models", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch models, status: %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			ID            string `json:"id"`
			Object        string `json:"object"`
			Created       int64  `json:"created"`
			OwnedBy       string `json:"owned_by"`
			ContextWindow int    `json:"context_window,omitempty"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode models response: %w", err)
	}

	models := make([]Model, 0, len(result.Data))
	for _, m := range result.Data {
		model := Model{
			ID:            m.ID,
			Provider:      "groq",
			Name:          m.ID,
			Status:        ModelStatusAvailable,
			ContextWindow: m.ContextWindow,
			CreatedAt:     m.Created,
		}

		// Try to match with static data
		if static, err := findStaticModel("groq", m.ID); err == nil {
			model.Name = static.Name
			model.Description = static.Description
			model.Capabilities = static.Capabilities
			model.PricingTier = static.PricingTier
			model.MaxTokens = static.MaxTokens
			model.InputPrice = static.InputPrice
			model.OutputPrice = static.OutputPrice
			model.SupportsVision = static.SupportsVision
			model.SupportsToolUse = static.SupportsToolUse
		}

		models = append(models, model)
	}

	return models, nil
}

func (p *groqProvider) ValidateAPIKey(ctx context.Context, apiKey string) error {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "GET", p.config.BaseURL+"/models", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate API key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid API key")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API key validation failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (p *groqProvider) GetDefaultModels() []Model {
	return GetStaticModels("groq")
}

// openRouterProvider implements ModelProvider for OpenRouter
type openRouterProvider struct {
	config ProviderConfig
}

// NewOpenRouterProvider creates a new OpenRouter provider
func NewOpenRouterProvider() ModelProvider {
	return &openRouterProvider{
		config: ProviderConfigs["openrouter"],
	}
}

func (p *openRouterProvider) Name() string {
	return "openrouter"
}

func (p *openRouterProvider) Config() ProviderConfig {
	return p.config
}

func (p *openRouterProvider) SupportsDynamicFetch() bool {
	return true
}

func (p *openRouterProvider) FetchModels(ctx context.Context, apiKey string) ([]Model, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://openrouter.ai/api/v1/models", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("HTTP-Referer", "https://gitscribe.ai") // Required by OpenRouter
	req.Header.Set("X-Title", "GitScribe")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch models, status: %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			ID            string `json:"id"`
			Name          string `json:"name"`
			Description   string `json:"description"`
			ContextLength int    `json:"context_length"`
			Pricing       struct {
				Prompt     float64 `json:"prompt"`
				Completion float64 `json:"completion"`
			} `json:"pricing"`
			TopProvider struct {
				ContextLength int  `json:"context_length"`
				IsModerated   bool `json:"is_moderated"`
			} `json:"top_provider"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode models response: %w", err)
	}

	models := make([]Model, 0, len(result.Data))
	for _, m := range result.Data {
		// Skip free models or those with zero pricing if desired
		// if m.Pricing.Prompt == 0 && m.Pricing.Completion == 0 {
		// 	continue
		// }

		model := Model{
			ID:              m.ID,
			Provider:        "openrouter",
			Name:            m.Name,
			Description:     m.Description,
			ContextWindow:   m.ContextLength,
			MaxTokens:       m.ContextLength,
			InputPrice:      m.Pricing.Prompt * 1000000, // Convert to per 1M
			OutputPrice:     m.Pricing.Completion * 1000000,
			Status:          ModelStatusAvailable,
			Capabilities:    []Capability{CapabilityChat},
			SupportsVision:  false, // Would need to check per model
			SupportsToolUse: false,
		}

		// Try to match with static data
		if static, err := findStaticModel("openrouter", m.ID); err == nil {
			model.Capabilities = static.Capabilities
			model.SupportsVision = static.SupportsVision
			model.SupportsToolUse = static.SupportsToolUse
			model.PricingTier = static.PricingTier
		}

		// Determine pricing tier based on prices
		if model.InputPrice == 0 && model.OutputPrice == 0 {
			model.PricingTier = PricingFree
		} else if model.InputPrice < 0.5 {
			model.PricingTier = PricingBudget
		} else if model.InputPrice < 5.0 {
			model.PricingTier = PricingStandard
		} else {
			model.PricingTier = PricingPremium
		}

		models = append(models, model)
	}

	return models, nil
}

func (p *openRouterProvider) ValidateAPIKey(ctx context.Context, apiKey string) error {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://openrouter.ai/api/v1/auth/key", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate API key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid API key")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API key validation failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (p *openRouterProvider) GetDefaultModels() []Model {
	return GetStaticModels("openrouter")
}

// ollamaProvider implements ModelProvider for Ollama (local)
type ollamaProvider struct {
	config ProviderConfig
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider() ModelProvider {
	return &ollamaProvider{
		config: ProviderConfigs["ollama"],
	}
}

func (p *ollamaProvider) Name() string {
	return "ollama"
}

func (p *ollamaProvider) Config() ProviderConfig {
	return p.config
}

func (p *ollamaProvider) SupportsDynamicFetch() bool {
	return true
}

func (p *ollamaProvider) FetchModels(ctx context.Context, apiKey string) ([]Model, error) {
	// Ollama doesn't use the OpenAI-compatible /models endpoint
	// It uses /api/tags for listing models
	client := &http.Client{Timeout: 10 * time.Second}

	// Try OpenAI-compatible endpoint first
	req, err := http.NewRequestWithContext(ctx, "GET", p.config.BaseURL+"/models", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		// Fall back to Ollama native API
		return p.fetchFromOllamaAPI(ctx)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return p.fetchFromOllamaAPI(ctx)
	}

	var result struct {
		Data []struct {
			ID     string `json:"id"`
			Object string `json:"object"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return p.fetchFromOllamaAPI(ctx)
	}

	models := make([]Model, 0, len(result.Data))
	for _, m := range result.Data {
		model := Model{
			ID:           m.ID,
			Provider:     "ollama",
			Name:         m.ID,
			Status:       ModelStatusAvailable,
			PricingTier:  PricingFree,
			Capabilities: []Capability{CapabilityChat},
		}

		// Try to match with static data
		if static, err := findStaticModel("ollama", m.ID); err == nil {
			model.Name = static.Name
			model.Description = static.Description
			model.Capabilities = static.Capabilities
			model.MaxTokens = static.MaxTokens
			model.ContextWindow = static.ContextWindow
		}

		models = append(models, model)
	}

	return models, nil
}

func (p *ollamaProvider) fetchFromOllamaAPI(ctx context.Context) ([]Model, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	// Extract base URL without /v1
	baseURL := "http://localhost:11434"

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama not running or unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama API returned status: %d", resp.StatusCode)
	}

	var result struct {
		Models []struct {
			Name       string `json:"name"`
			Model      string `json:"model"`
			ModifiedAt string `json:"modified_at"`
			Size       int64  `json:"size"`
			Details    struct {
				Format            string   `json:"format"`
				Family            string   `json:"family"`
				Families          []string `json:"families"`
				ParameterSize     string   `json:"parameter_size"`
				QuantizationLevel string   `json:"quantization_level"`
			} `json:"details"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode ollama response: %w", err)
	}

	models := make([]Model, 0, len(result.Models))
	for _, m := range result.Models {
		model := Model{
			ID:              m.Name,
			Provider:        "ollama",
			Name:            m.Name,
			Status:          ModelStatusAvailable,
			PricingTier:     PricingFree,
			Capabilities:    []Capability{CapabilityChat},
			SupportsVision:  false,
			SupportsToolUse: false,
		}

		// Add details from Ollama
		if m.Details.Family != "" {
			model.Description = fmt.Sprintf("%s %s model", m.Details.Family, m.Details.ParameterSize)
		}

		// Try to match with static data for richer info
		baseName := m.Name
		if idx := len(m.Name); idx > 0 {
			// Remove :latest or :version tag
			for i := len(m.Name) - 1; i >= 0; i-- {
				if m.Name[i] == ':' {
					baseName = m.Name[:i]
					break
				}
			}
		}

		if static, err := findStaticModel("ollama", baseName); err == nil {
			model.Name = static.Name
			model.Capabilities = static.Capabilities
			model.MaxTokens = static.MaxTokens
			model.ContextWindow = static.ContextWindow
			model.SupportsVision = static.SupportsVision
			model.SupportsToolUse = static.SupportsToolUse
		}

		models = append(models, model)
	}

	return models, nil
}

func (p *ollamaProvider) ValidateAPIKey(ctx context.Context, apiKey string) error {
	// Ollama doesn't require API keys for local usage
	// Just check if Ollama is running
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:11434/api/tags", nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ollama not running at localhost:11434: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	return nil
}

func (p *ollamaProvider) GetDefaultModels() []Model {
	return GetStaticModels("ollama")
}

// findStaticModel looks up a model in the static catalog
func findStaticModel(provider, id string) (*Model, error) {
	models := GetStaticModels(provider)
	for i := range models {
		if models[i].ID == id {
			return &models[i], nil
		}
		// Check aliases
		for _, alias := range models[i].Aliases {
			if alias == id {
				return &models[i], nil
			}
		}
	}
	return nil, fmt.Errorf("model not found in static catalog: %s", id)
}

// Model needs a CreatedAt field for the provider fetchers
// Add it to the struct
func init() {
	// This will be called at package init time
	// The CreatedAt field was added to Model struct in models.go
}
