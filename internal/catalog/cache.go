package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// DefaultCacheDuration is how long to use cached data before refreshing
	DefaultCacheDuration = 24 * time.Hour

	// CacheFileName is the name of the cache file
	CacheFileName = "model-catalog-cache.json"

	// MinRefreshInterval is the minimum time between refresh attempts
	MinRefreshInterval = 1 * time.Hour
)

// Cache stores the model catalog with metadata
type Cache struct {
	Catalog   ModelCatalog         `json:"catalog"`
	FetchedAt map[string]time.Time `json:"fetched_at"` // per provider
	Version   string               `json:"version"`
}

// CacheManager handles caching operations
type CacheManager struct {
	cacheDir           string
	cacheFile          string
	cacheDuration      time.Duration
	minRefreshInterval time.Duration
}

// CacheOptions configures the cache manager
type CacheOptions struct {
	CacheDir           string
	CacheDuration      time.Duration
	MinRefreshInterval time.Duration
}

// NewCacheManager creates a new cache manager
func NewCacheManager(opts CacheOptions) (*CacheManager, error) {
	if opts.CacheDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		opts.CacheDir = filepath.Join(home, ".multiagent", "cache")
	}

	if opts.CacheDuration == 0 {
		opts.CacheDuration = DefaultCacheDuration
	}

	if opts.MinRefreshInterval == 0 {
		opts.MinRefreshInterval = MinRefreshInterval
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(opts.CacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &CacheManager{
		cacheDir:           opts.CacheDir,
		cacheFile:          filepath.Join(opts.CacheDir, CacheFileName),
		cacheDuration:      opts.CacheDuration,
		minRefreshInterval: opts.MinRefreshInterval,
	}, nil
}

// Load reads the cache from disk
func (cm *CacheManager) Load() (*Cache, error) {
	data, err := os.ReadFile(cm.cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty cache
			return &Cache{
				Catalog: ModelCatalog{
					Metadata: CatalogMetadata{
						Version: "1.0",
						Schema:  "model-catalog-v1",
					},
					Providers: []ProviderModels{},
				},
				FetchedAt: make(map[string]time.Time),
				Version:   "1.0",
			}, nil
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to parse cache file: %w", err)
	}

	if cache.FetchedAt == nil {
		cache.FetchedAt = make(map[string]time.Time)
	}

	return &cache, nil
}

// Save writes the cache to disk
func (cm *CacheManager) Save(cache *Cache) error {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	// Write with restricted permissions
	if err := os.WriteFile(cm.cacheFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// IsStale checks if a provider's cache entry is stale
func (cm *CacheManager) IsStale(cache *Cache, provider string) bool {
	fetchedAt, ok := cache.FetchedAt[provider]
	if !ok {
		return true
	}

	return time.Since(fetchedAt) > cm.cacheDuration
}

// CanRefresh checks if enough time has passed to attempt a refresh
func (cm *CacheManager) CanRefresh(cache *Cache, provider string) bool {
	fetchedAt, ok := cache.FetchedAt[provider]
	if !ok {
		return true
	}

	return time.Since(fetchedAt) > cm.minRefreshInterval
}

// UpdateProvider updates the cache with new provider models
func (cm *CacheManager) UpdateProvider(cache *Cache, provider string, models []Model) {
	// Find existing provider entry
	var found bool
	for i := range cache.Catalog.Providers {
		if cache.Catalog.Providers[i].Provider.Name == provider {
			cache.Catalog.Providers[i].Models = models
			cache.Catalog.Providers[i].Updated = time.Now()
			found = true
			break
		}
	}

	// Add new provider entry if not found
	if !found {
		if config, ok := GetProviderConfig(provider); ok {
			cache.Catalog.Providers = append(cache.Catalog.Providers, ProviderModels{
				Provider: config,
				Models:   models,
				Updated:  time.Now(),
			})
		}
	}

	// Update fetch timestamp
	cache.FetchedAt[provider] = time.Now()

	// Update catalog metadata
	cache.Catalog.Metadata.LastUpdated = time.Now()
}

// Clear removes the entire cache
func (cm *CacheManager) Clear() error {
	if err := os.Remove(cm.cacheFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	return nil
}

// GetCacheAge returns how old the cache is for a specific provider
func (cm *CacheManager) GetCacheAge(cache *Cache, provider string) time.Duration {
	fetchedAt, ok := cache.FetchedAt[provider]
	if !ok {
		return time.Duration(1<<63 - 1) // Max duration
	}
	return time.Since(fetchedAt)
}

// WarmCache pre-populates the cache with static models
func (cm *CacheManager) WarmCache() (*Cache, error) {
	cache, err := cm.Load()
	if err != nil {
		return nil, err
	}

	// Add static models for all providers
	for providerName := range StaticModels {
		models := GetStaticModels(providerName)
		if len(models) > 0 {
			cm.UpdateProvider(cache, providerName, models)
		}
	}

	// Save the warmed cache
	if err := cm.Save(cache); err != nil {
		return nil, err
	}

	return cache, nil
}
