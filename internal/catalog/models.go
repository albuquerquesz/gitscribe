package catalog

import (
	"encoding/json"
	"fmt"
	"time"
)


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


type PricingTier string

const (
	PricingFree       PricingTier = "free"
	PricingBudget     PricingTier = "budget"
	PricingStandard   PricingTier = "standard"
	PricingPremium    PricingTier = "premium"
	PricingEnterprise PricingTier = "enterprise"
)


type ModelStatus string

const (
	ModelStatusAvailable ModelStatus = "available"
	ModelStatusPreview   ModelStatus = "preview"
	ModelStatusBeta      ModelStatus = "beta"
	ModelStatusLegacy    ModelStatus = "legacy"
	ModelStatusRemoved   ModelStatus = "removed"
)


type Model struct {
	ID              string       `json:"id" yaml:"id"`
	Provider        string       `json:"provider" yaml:"provider"`
	Name            string       `json:"name" yaml:"name"`
	Description     string       `json:"description,omitempty" yaml:"description,omitempty"`
	MaxTokens       int          `json:"max_tokens" yaml:"max_tokens"`
	InputPrice      float64      `json:"input_price_per_1m" yaml:"input_price_per_1m"`   
	OutputPrice     float64      `json:"output_price_per_1m" yaml:"output_price_per_1m"` 
	PricingTier     PricingTier  `json:"pricing_tier" yaml:"pricing_tier"`
	Capabilities    []Capability `json:"capabilities" yaml:"capabilities"`
	Status          ModelStatus  `json:"status" yaml:"status"`
	ContextWindow   int          `json:"context_window" yaml:"context_window"`
	TrainingCutoff  string       `json:"training_cutoff,omitempty" yaml:"training_cutoff,omitempty"`
	SupportsVision  bool         `json:"supports_vision" yaml:"supports_vision"`
	SupportsToolUse bool         `json:"supports_tool_use" yaml:"supports_tool_use"`
	RecommendedFor  []string     `json:"recommended_for,omitempty" yaml:"recommended_for,omitempty"`
	Aliases         []string     `json:"aliases,omitempty" yaml:"aliases,omitempty"`
	CreatedAt       int64        `json:"created_at,omitempty" yaml:"created_at,omitempty"` 
}


type ProviderConfig struct {
	Name           string            `json:"name" yaml:"name"`
	BaseURL        string            `json:"base_url" yaml:"base_url"`
	AuthMethod     AuthMethod        `json:"auth_method" yaml:"auth_method"`
	DefaultHeaders map[string]string `json:"default_headers,omitempty" yaml:"default_headers,omitempty"`
	ModelsEndpoint string            `json:"models_endpoint,omitempty" yaml:"models_endpoint,omitempty"`
	SupportsList   bool              `json:"supports_list" yaml:"supports_list"` 
	RequiresAuth   bool              `json:"requires_auth" yaml:"requires_auth"`
	RateLimitRPS   int               `json:"rate_limit_rps,omitempty" yaml:"rate_limit_rps,omitempty"`
}


type AuthMethod string

const (
	AuthMethodAPIKey AuthMethod = "api_key"
	AuthMethodBearer AuthMethod = "bearer"
	AuthMethodBasic  AuthMethod = "basic"
	AuthMethodNone   AuthMethod = "none"
	AuthMethodCustom AuthMethod = "custom"
)


type ProviderModels struct {
	Provider ProviderConfig `json:"provider" yaml:"provider"`
	Models   []Model        `json:"models" yaml:"models"`
	Updated  time.Time      `json:"updated" yaml:"updated"`
}


type CatalogMetadata struct {
	Version     string    `json:"version" yaml:"version"`
	Generated   time.Time `json:"generated" yaml:"generated"`
	Schema      string    `json:"schema" yaml:"schema"`
	LastUpdated time.Time `json:"last_updated" yaml:"last_updated"`
}


type ModelCatalog struct {
	Metadata  CatalogMetadata  `json:"metadata" yaml:"metadata"`
	Providers []ProviderModels `json:"providers" yaml:"providers"`
}


type FilterOptions struct {
	Provider       string
	Capability     Capability
	PricingTier    PricingTier
	Status         ModelStatus
	MinContextSize int
	MaxPrice       float64
	SupportsVision bool
}


func (c *ModelCatalog) GetModelByID(id string) (*Model, error) {
	for _, pm := range c.Providers {
		for i := range pm.Models {
			if pm.Models[i].ID == id {
				return &pm.Models[i], nil
			}
			
			for _, alias := range pm.Models[i].Aliases {
				if alias == id {
					return &pm.Models[i], nil
				}
			}
		}
	}
	return nil, fmt.Errorf("model not found: %s", id)
}


func (c *ModelCatalog) GetModelsByProvider(provider string) []Model {
	for _, pm := range c.Providers {
		if pm.Provider.Name == provider {
			return pm.Models
		}
	}
	return nil
}


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


func (m Model) String() string {
	return fmt.Sprintf("%s (%s) - %s - %d tokens - $%.2f/1M in",
		m.Name, m.ID, m.PricingTier, m.ContextWindow, m.InputPrice)
}


func (m Model) JSON() ([]byte, error) {
	return json.MarshalIndent(m, "", "  ")
}


func (m Model) IsAvailable() bool {
	return m.Status == ModelStatusAvailable || m.Status == ModelStatusPreview || m.Status == ModelStatusBeta
}
