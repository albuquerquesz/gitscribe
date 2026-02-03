package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)


type anthropicProvider struct {
	config ProviderConfig
}


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
	return false 
}

func (p *anthropicProvider) FetchModels(ctx context.Context, apiKey string) ([]Model, error) {
	
	return nil, fmt.Errorf("anthropic does not support dynamic model fetching")
}

func (p *anthropicProvider) ValidateAPIKey(ctx context.Context, apiKey string) error {
	
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


type openAIProvider struct {
	config ProviderConfig
}


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


type groqProvider struct {
	config ProviderConfig
}


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


type openRouterProvider struct {
	config ProviderConfig
}


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
	req.Header.Set("HTTP-Referer", "https://gitscribe.ai") 
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
		
		
		
		

		model := Model{
			ID:              m.ID,
			Provider:        "openrouter",
			Name:            m.Name,
			Description:     m.Description,
			ContextWindow:   m.ContextLength,
			MaxTokens:       m.ContextLength,
			InputPrice:      m.Pricing.Prompt * 1000000, 
			OutputPrice:     m.Pricing.Completion * 1000000,
			Status:          ModelStatusAvailable,
			Capabilities:    []Capability{CapabilityChat},
			SupportsVision:  false, 
			SupportsToolUse: false,
		}

		
		if static, err := findStaticModel("openrouter", m.ID); err == nil {
			model.Capabilities = static.Capabilities
			model.SupportsVision = static.SupportsVision
			model.SupportsToolUse = static.SupportsToolUse
			model.PricingTier = static.PricingTier
		}

		
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


type ollamaProvider struct {
	config ProviderConfig
}


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
	
	
	client := &http.Client{Timeout: 10 * time.Second}

	
	req, err := http.NewRequestWithContext(ctx, "GET", p.config.BaseURL+"/models", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		
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

		
		if m.Details.Family != "" {
			model.Description = fmt.Sprintf("%s %s model", m.Details.Family, m.Details.ParameterSize)
		}

		
		baseName := m.Name
		if idx := len(m.Name); idx > 0 {
			
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


func findStaticModel(provider, id string) (*Model, error) {
	models := GetStaticModels(provider)
	for i := range models {
		if models[i].ID == id {
			return &models[i], nil
		}
		
		for _, alias := range models[i].Aliases {
			if alias == id {
				return &models[i], nil
			}
		}
	}
	return nil, fmt.Errorf("model not found in static catalog: %s", id)
}



func init() {
	
	
}
