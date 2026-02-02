package catalog

import (
	"encoding/json"
	"fmt"
	"time"
)

// Capability represents what a model can do
type Capability string

const (
	CapabilityChat         Capability = "chat"
	CapabilityCompletion   Capability = "completion"
	CapabilityVision       Capability = "vision"
	CapabilityEmbedding    Capability = "embedding"
	CapabilityReasoning    Capability = "reasoning"
	CapabilityCode         Capability = "code"
	CapabilityFunctionCall Capability = "function_calling"
	CapabilityJSON         Capability = "json_mode"
)

// PricingTier indicates the cost category
type PricingTier string

const (
	PricingFree       PricingTier = "free"
	PricingBudget     PricingTier = "budget"
	PricingStandard   PricingTier = "standard"
	PricingPremium    PricingTier = "premium"
	PricingEnterprise PricingTier = "enterprise"
)

// ModelStatus indicates availability status
type ModelStatus string

const (
	ModelStatusAvailable ModelStatus = "available"
	ModelStatusPreview   ModelStatus = "preview"
	ModelStatusBeta      ModelStatus = "beta"
	ModelStatusLegacy    ModelStatus = "legacy"
	ModelStatusRemoved   ModelStatus = "removed"
)

// Model represents a single AI model with its metadata
type Model struct {
	ID              string       `json:"id" yaml:"id"`
	Provider        string       `json:"provider" yaml:"provider"`
	Name            string       `json:"name" yaml:"name"`
	Description     string       `json:"description,omitempty" yaml:"description,omitempty"`
	MaxTokens       int          `json:"max_tokens" yaml:"max_tokens"`
	InputPrice      float64      `json:"input_price_per_1m" yaml:"input_price_per_1m"`   // per 1M tokens
	OutputPrice     float64      `json:"output_price_per_1m" yaml:"output_price_per_1m"` // per 1M tokens
	PricingTier     PricingTier  `json:"pricing_tier" yaml:"pricing_tier"`
	Capabilities    []Capability `json:"capabilities" yaml:"capabilities"`
	Status          ModelStatus  `json:"status" yaml:"status"`
	ContextWindow   int          `json:"context_window" yaml:"context_window"`
	TrainingCutoff  string       `json:"training_cutoff,omitempty" yaml:"training_cutoff,omitempty"`
	SupportsVision  bool         `json:"supports_vision" yaml:"supports_vision"`
	SupportsToolUse bool         `json:"supports_tool_use" yaml:"supports_tool_use"`
	RecommendedFor  []string     `json:"recommended_for,omitempty" yaml:"recommended_for,omitempty"`
	Aliases         []string     `json:"aliases,omitempty" yaml:"aliases,omitempty"`
	CreatedAt       int64        `json:"created_at,omitempty" yaml:"created_at,omitempty"` // Unix timestamp from API
}

// ProviderConfig contains provider-specific settings
type ProviderConfig struct {
	Name           string            `json:"name" yaml:"name"`
	BaseURL        string            `json:"base_url" yaml:"base_url"`
	AuthMethod     AuthMethod        `json:"auth_method" yaml:"auth_method"`
	DefaultHeaders map[string]string `json:"default_headers,omitempty" yaml:"default_headers,omitempty"`
	ModelsEndpoint string            `json:"models_endpoint,omitempty" yaml:"models_endpoint,omitempty"`
	SupportsList   bool              `json:"supports_list" yaml:"supports_list"` // Can fetch models dynamically
	RequiresAuth   bool              `json:"requires_auth" yaml:"requires_auth"`
	RateLimitRPS   int               `json:"rate_limit_rps,omitempty" yaml:"rate_limit_rps,omitempty"`
}

// AuthMethod defines how authentication is handled
type AuthMethod string

const (
	AuthMethodAPIKey AuthMethod = "api_key"
	AuthMethodBearer AuthMethod = "bearer"
	AuthMethodBasic  AuthMethod = "basic"
	AuthMethodNone   AuthMethod = "none"
	AuthMethodCustom AuthMethod = "custom"
)

// ProviderModels groups models by provider
type ProviderModels struct {
	Provider ProviderConfig `json:"provider" yaml:"provider"`
	Models   []Model        `json:"models" yaml:"models"`
	Updated  time.Time      `json:"updated" yaml:"updated"`
}

// CatalogMetadata tracks catalog version and freshness
type CatalogMetadata struct {
	Version     string    `json:"version" yaml:"version"`
	Generated   time.Time `json:"generated" yaml:"generated"`
	Schema      string    `json:"schema" yaml:"schema"`
	LastUpdated time.Time `json:"last_updated" yaml:"last_updated"`
}

// ModelCatalog is the root structure containing all models
type ModelCatalog struct {
	Metadata  CatalogMetadata  `json:"metadata" yaml:"metadata"`
	Providers []ProviderModels `json:"providers" yaml:"providers"`
}

// FilterOptions for searching models
type FilterOptions struct {
	Provider       string
	Capability     Capability
	PricingTier    PricingTier
	Status         ModelStatus
	MinContextSize int
	MaxPrice       float64
	SupportsVision bool
}

// GetModelByID finds a model by its ID across all providers
func (c *ModelCatalog) GetModelByID(id string) (*Model, error) {
	for _, pm := range c.Providers {
		for i := range pm.Models {
			if pm.Models[i].ID == id {
				return &pm.Models[i], nil
			}
			// Check aliases
			for _, alias := range pm.Models[i].Aliases {
				if alias == id {
					return &pm.Models[i], nil
				}
			}
		}
	}
	return nil, fmt.Errorf("model not found: %s", id)
}

// GetModelsByProvider returns all models for a specific provider
func (c *ModelCatalog) GetModelsByProvider(provider string) []Model {
	for _, pm := range c.Providers {
		if pm.Provider.Name == provider {
			return pm.Models
		}
	}
	return nil
}

// Filter returns models matching the given criteria
func (c *ModelCatalog) Filter(opts FilterOptions) []Model {
	var results []Model

	for _, pm := range c.Providers {
		if opts.Provider != "" && pm.Provider.Name != opts.Provider {
			continue
		}

		for _, model := range pm.Models {
			if !c.matchesFilter(model, opts) {
				continue
			}
			results = append(results, model)
		}
	}

	return results
}

func (c *ModelCatalog) matchesFilter(model Model, opts FilterOptions) bool {
	if opts.Capability != "" && !hasCapability(model.Capabilities, opts.Capability) {
		return false
	}
	if opts.PricingTier != "" && model.PricingTier != opts.PricingTier {
		return false
	}
	if opts.Status != "" && model.Status != opts.Status {
		return false
	}
	if opts.MinContextSize > 0 && model.ContextWindow < opts.MinContextSize {
		return false
	}
	if opts.MaxPrice > 0 && model.InputPrice > opts.MaxPrice {
		return false
	}
	if opts.SupportsVision && !model.SupportsVision {
		return false
	}
	return true
}

func hasCapability(caps []Capability, target Capability) bool {
	for _, c := range caps {
		if c == target {
			return true
		}
	}
	return false
}

// String returns a formatted display string for the model
func (m Model) String() string {
	return fmt.Sprintf("%s (%s) - %s - %d tokens - $%.2f/1M in",
		m.Name, m.ID, m.PricingTier, m.ContextWindow, m.InputPrice)
}

// JSON returns JSON representation
func (m Model) JSON() ([]byte, error) {
	return json.MarshalIndent(m, "", "  ")
}

// IsAvailable returns true if the model is currently usable
func (m Model) IsAvailable() bool {
	return m.Status == ModelStatusAvailable || m.Status == ModelStatusPreview || m.Status == ModelStatusBeta
}
