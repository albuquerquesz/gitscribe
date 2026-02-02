package catalog

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CatalogManager is the main interface for working with the model catalog
type CatalogManager struct {
	factory      *ProviderFactory
	cacheManager *CacheManager
	cache        *Cache
	mu           sync.RWMutex

	// API key resolver - should be set to fetch keys from keyring
	apiKeyResolver func(provider string) (string, error)
}

// ManagerOptions for configuring the catalog manager
type ManagerOptions struct {
	CacheOptions   CacheOptions
	APIKeyResolver func(provider string) (string, error)
}

// NewCatalogManager creates a new catalog manager
func NewCatalogManager(opts ManagerOptions) (*CatalogManager, error) {
	cacheManager, err := NewCacheManager(opts.CacheOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache manager: %w", err)
	}

	// Load or warm cache
	cache, err := cacheManager.Load()
	if err != nil {
		// Try to warm cache with static data
		cache, err = cacheManager.WarmCache()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize cache: %w", err)
		}
	}

	return &CatalogManager{
		factory:        NewProviderFactory(),
		cacheManager:   cacheManager,
		cache:          cache,
		apiKeyResolver: opts.APIKeyResolver,
	}, nil
}

// GetModel retrieves a model by ID from the catalog
func (cm *CatalogManager) GetModel(id string) (*Model, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.cache.Catalog.GetModelByID(id)
}

// GetModelsByProvider returns all models for a provider
func (cm *CatalogManager) GetModelsByProvider(provider string) ([]Model, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	models := cm.cache.Catalog.GetModelsByProvider(provider)
	if models == nil {
		// Try to get static models as fallback
		models = GetStaticModels(provider)
		if models == nil {
			return nil, fmt.Errorf("provider not found: %s", provider)
		}
	}

	return models, nil
}

// ListProviders returns all available providers
func (cm *CatalogManager) ListProviders() []string {
	return cm.factory.List()
}

// GetProviderConfig returns configuration for a provider
func (cm *CatalogManager) GetProviderConfig(name string) (*ProviderConfig, error) {
	config, ok := GetProviderConfig(name)
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return &config, nil
}

// FilterModels returns models matching the criteria
func (cm *CatalogManager) FilterModels(opts FilterOptions) []Model {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.cache.Catalog.Filter(opts)
}

// RefreshProvider updates the catalog for a specific provider
func (cm *CatalogManager) RefreshProvider(ctx context.Context, provider string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if we can refresh (rate limiting)
	if !cm.cacheManager.CanRefresh(cm.cache, provider) {
		return fmt.Errorf("cannot refresh %s: minimum refresh interval not met", provider)
	}

	// Get the provider
	p, err := cm.factory.Get(provider)
	if err != nil {
		return err
	}

	// Check if dynamic fetching is supported
	if !p.SupportsDynamicFetch() {
		// Just update with static models
		models := p.GetDefaultModels()
		cm.cacheManager.UpdateProvider(cm.cache, provider, models)
		return cm.cacheManager.Save(cm.cache)
	}

	// Get API key
	if cm.apiKeyResolver == nil {
		return fmt.Errorf("API key resolver not configured")
	}

	apiKey, err := cm.apiKeyResolver(provider)
	if err != nil {
		return fmt.Errorf("failed to get API key for %s: %w", provider, err)
	}

	// Fetch models from API
	models, err := p.FetchModels(ctx, apiKey)
	if err != nil {
		// Fall back to static models on error
		models = p.GetDefaultModels()
		// Still update the cache to mark attempt time
	}

	// Update cache
	cm.cacheManager.UpdateProvider(cm.cache, provider, models)

	// Save cache
	if err := cm.cacheManager.Save(cm.cache); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	return nil
}

// RefreshAll updates the catalog for all providers that support dynamic fetching
func (cm *CatalogManager) RefreshAll(ctx context.Context) error {
	providers := cm.factory.List()

	var lastErr error
	for _, provider := range providers {
		p, err := cm.factory.Get(provider)
		if err != nil {
			continue
		}

		if !p.SupportsDynamicFetch() {
			continue
		}

		if err := cm.RefreshProvider(ctx, provider); err != nil {
			lastErr = err
			// Continue with other providers
		}
	}

	return lastErr
}

// GetCacheStatus returns information about cache freshness
func (cm *CatalogManager) GetCacheStatus() map[string]CacheStatus {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	status := make(map[string]CacheStatus)
	providers := cm.factory.List()

	for _, provider := range providers {
		age := cm.cacheManager.GetCacheAge(cm.cache, provider)
		status[provider] = CacheStatus{
			Age:         age,
			IsStale:     cm.cacheManager.IsStale(cm.cache, provider),
			CanRefresh:  cm.cacheManager.CanRefresh(cm.cache, provider),
			LastFetched: cm.cache.FetchedAt[provider],
		}
	}

	return status
}

// CacheStatus provides information about a provider's cache state
type CacheStatus struct {
	Age         time.Duration `json:"age"`
	IsStale     bool          `json:"is_stale"`
	CanRefresh  bool          `json:"can_refresh"`
	LastFetched time.Time     `json:"last_fetched,omitempty"`
}

// ClearCache removes all cached data
func (cm *CatalogManager) ClearCache() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	return cm.cacheManager.Clear()
}

// GetRecommendedModels returns models recommended for specific use cases
func (cm *CatalogManager) GetRecommendedModels(useCase string) []Model {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var results []Model

	for _, pm := range cm.cache.Catalog.Providers {
		for _, model := range pm.Models {
			if !model.IsAvailable() {
				continue
			}

			for _, rec := range model.RecommendedFor {
				if rec == useCase {
					results = append(results, model)
					break
				}
			}
		}
	}

	return results
}

// SuggestModel recommends a model based on requirements
func (cm *CatalogManager) SuggestModel(requirements ModelRequirements) (*Model, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Score models based on requirements
	var bestModel *Model
	bestScore := -1

	for _, pm := range cm.cache.Catalog.Providers {
		for i := range pm.Models {
			model := &pm.Models[i]

			if !model.IsAvailable() {
				continue
			}

			score := cm.scoreModel(model, requirements)
			if score > bestScore {
				bestScore = score
				bestModel = model
			}
		}
	}

	if bestModel == nil {
		return nil, fmt.Errorf("no suitable model found")
	}

	return bestModel, nil
}

// ModelRequirements for model selection
type ModelRequirements struct {
	Provider       string
	MinContextSize int
	MaxPrice       float64
	RequiresVision bool
	RequiresTools  bool
	Capabilities   []Capability
}

func (cm *CatalogManager) scoreModel(model *Model, req ModelRequirements) int {
	score := 0

	// Provider preference
	if req.Provider != "" && model.Provider == req.Provider {
		score += 50
	}

	// Context window
	if req.MinContextSize > 0 {
		if model.ContextWindow >= req.MinContextSize {
			score += 20
		} else {
			return -1 // Hard requirement
		}
	}

	// Price
	if req.MaxPrice > 0 {
		if model.InputPrice <= req.MaxPrice {
			score += int((req.MaxPrice - model.InputPrice) * 10)
		} else {
			return -1 // Hard requirement
		}
	}

	// Capabilities
	for _, cap := range req.Capabilities {
		if hasCapability(model.Capabilities, cap) {
			score += 10
		} else {
			return -1 // Hard requirement
		}
	}

	// Vision
	if req.RequiresVision && !model.SupportsVision {
		return -1
	}
	if model.SupportsVision {
		score += 5
	}

	// Tools
	if req.RequiresTools && !model.SupportsToolUse {
		return -1
	}
	if model.SupportsToolUse {
		score += 5
	}

	// Prefer newer models (higher created timestamp)
	if model.CreatedAt > 0 {
		score += int(model.CreatedAt / 1e9 / 86400) // Days since epoch
	}

	return score
}

// ValidateAPIKey checks if an API key is valid for a provider
func (cm *CatalogManager) ValidateAPIKey(ctx context.Context, provider, apiKey string) error {
	p, err := cm.factory.Get(provider)
	if err != nil {
		return err
	}

	return p.ValidateAPIKey(ctx, apiKey)
}

// ForceRefresh bypasses rate limiting and refreshes a provider
func (cm *CatalogManager) ForceRefresh(ctx context.Context, provider string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Get the provider
	p, err := cm.factory.Get(provider)
	if err != nil {
		return err
	}

	if !p.SupportsDynamicFetch() {
		models := p.GetDefaultModels()
		cm.cacheManager.UpdateProvider(cm.cache, provider, models)
		return cm.cacheManager.Save(cm.cache)
	}

	// Get API key
	if cm.apiKeyResolver == nil {
		return fmt.Errorf("API key resolver not configured")
	}

	apiKey, err := cm.apiKeyResolver(provider)
	if err != nil {
		return fmt.Errorf("failed to get API key for %s: %w", provider, err)
	}

	// Fetch models from API
	models, err := p.FetchModels(ctx, apiKey)
	if err != nil {
		return fmt.Errorf("failed to fetch models: %w", err)
	}

	// Update cache
	cm.cacheManager.UpdateProvider(cm.cache, provider, models)

	return cm.cacheManager.Save(cm.cache)
}

// GetCatalog returns the full catalog
func (cm *CatalogManager) GetCatalog() ModelCatalog {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Return a copy
	catalog := cm.cache.Catalog
	return catalog
}
