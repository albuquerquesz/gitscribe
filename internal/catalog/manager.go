package catalog

import (
	"context"
	"fmt"
	"sync"
	"time"
)


type CatalogManager struct {
	factory      *ProviderFactory
	cacheManager *CacheManager
	cache        *Cache
	mu           sync.RWMutex

	
	apiKeyResolver func(provider string) (string, error)
}


type ManagerOptions struct {
	CacheOptions   CacheOptions
	APIKeyResolver func(provider string) (string, error)
}


func NewCatalogManager(opts ManagerOptions) (*CatalogManager, error) {
	cacheManager, err := NewCacheManager(opts.CacheOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache manager: %w", err)
	}

	
	cache, err := cacheManager.Load()
	if err != nil {
		
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


func (cm *CatalogManager) GetModel(id string) (*Model, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.cache.Catalog.GetModelByID(id)
}


func (cm *CatalogManager) GetModelsByProvider(provider string) ([]Model, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	models := cm.cache.Catalog.GetModelsByProvider(provider)
	if models == nil {
		
		models = GetStaticModels(provider)
		if models == nil {
			return nil, fmt.Errorf("provider not found: %s", provider)
		}
	}

	return models, nil
}


func (cm *CatalogManager) ListProviders() []string {
	return cm.factory.List()
}


func (cm *CatalogManager) GetProviderConfig(name string) (*ProviderConfig, error) {
	config, ok := GetProviderConfig(name)
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return &config, nil
}


func (cm *CatalogManager) FilterModels(opts FilterOptions) []Model {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.cache.Catalog.Filter(opts)
}


func (cm *CatalogManager) RefreshProvider(ctx context.Context, provider string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	
	if !cm.cacheManager.CanRefresh(cm.cache, provider) {
		return fmt.Errorf("cannot refresh %s: minimum refresh interval not met", provider)
	}

	
	p, err := cm.factory.Get(provider)
	if err != nil {
		return err
	}

	
	if !p.SupportsDynamicFetch() {
		
		models := p.GetDefaultModels()
		cm.cacheManager.UpdateProvider(cm.cache, provider, models)
		return cm.cacheManager.Save(cm.cache)
	}

	
	if cm.apiKeyResolver == nil {
		return fmt.Errorf("API key resolver not configured")
	}

	apiKey, err := cm.apiKeyResolver(provider)
	if err != nil {
		return fmt.Errorf("failed to get API key for %s: %w", provider, err)
	}

	
	models, err := p.FetchModels(ctx, apiKey)
	if err != nil {
		
		models = p.GetDefaultModels()
		
	}

	
	cm.cacheManager.UpdateProvider(cm.cache, provider, models)

	
	if err := cm.cacheManager.Save(cm.cache); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	return nil
}


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
			
		}
	}

	return lastErr
}


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


type CacheStatus struct {
	Age         time.Duration `json:"age"`
	IsStale     bool          `json:"is_stale"`
	CanRefresh  bool          `json:"can_refresh"`
	LastFetched time.Time     `json:"last_fetched,omitempty"`
}


func (cm *CatalogManager) ClearCache() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	return cm.cacheManager.Clear()
}


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


func (cm *CatalogManager) SuggestModel(requirements ModelRequirements) (*Model, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	
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

	
	if req.Provider != "" && model.Provider == req.Provider {
		score += 50
	}

	
	if req.MinContextSize > 0 {
		if model.ContextWindow >= req.MinContextSize {
			score += 20
		} else {
			return -1 
		}
	}

	
	if req.MaxPrice > 0 {
		if model.InputPrice <= req.MaxPrice {
			score += int((req.MaxPrice - model.InputPrice) * 10)
		} else {
			return -1 
		}
	}

	
	for _, cap := range req.Capabilities {
		if hasCapability(model.Capabilities, cap) {
			score += 10
		} else {
			return -1 
		}
	}

	
	if req.RequiresVision && !model.SupportsVision {
		return -1
	}
	if model.SupportsVision {
		score += 5
	}

	
	if req.RequiresTools && !model.SupportsToolUse {
		return -1
	}
	if model.SupportsToolUse {
		score += 5
	}

	
	if model.CreatedAt > 0 {
		score += int(model.CreatedAt / 1e9 / 86400) 
	}

	return score
}


func (cm *CatalogManager) ValidateAPIKey(ctx context.Context, provider, apiKey string) error {
	p, err := cm.factory.Get(provider)
	if err != nil {
		return err
	}

	return p.ValidateAPIKey(ctx, apiKey)
}


func (cm *CatalogManager) ForceRefresh(ctx context.Context, provider string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	
	p, err := cm.factory.Get(provider)
	if err != nil {
		return err
	}

	if !p.SupportsDynamicFetch() {
		models := p.GetDefaultModels()
		cm.cacheManager.UpdateProvider(cm.cache, provider, models)
		return cm.cacheManager.Save(cm.cache)
	}

	
	if cm.apiKeyResolver == nil {
		return fmt.Errorf("API key resolver not configured")
	}

	apiKey, err := cm.apiKeyResolver(provider)
	if err != nil {
		return fmt.Errorf("failed to get API key for %s: %w", provider, err)
	}

	
	models, err := p.FetchModels(ctx, apiKey)
	if err != nil {
		return fmt.Errorf("failed to fetch models: %w", err)
	}

	
	cm.cacheManager.UpdateProvider(cm.cache, provider, models)

	return cm.cacheManager.Save(cm.cache)
}


func (cm *CatalogManager) GetCatalog() ModelCatalog {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	
	catalog := cm.cache.Catalog
	return catalog
}
