package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	
	DefaultCacheDuration = 24 * time.Hour

	
	CacheFileName = "model-catalog-cache.json"

	
	MinRefreshInterval = 1 * time.Hour
)


type Cache struct {
	Catalog   ModelCatalog         `json:"catalog"`
	FetchedAt map[string]time.Time `json:"fetched_at"` 
	Version   string               `json:"version"`
}


type CacheManager struct {
	cacheDir           string
	cacheFile          string
	cacheDuration      time.Duration
	minRefreshInterval time.Duration
}


type CacheOptions struct {
	CacheDir           string
	CacheDuration      time.Duration
	MinRefreshInterval time.Duration
}


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


func (cm *CacheManager) Load() (*Cache, error) {
	data, err := os.ReadFile(cm.cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			
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


func (cm *CacheManager) Save(cache *Cache) error {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	
	if err := os.WriteFile(cm.cacheFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}


func (cm *CacheManager) IsStale(cache *Cache, provider string) bool {
	fetchedAt, ok := cache.FetchedAt[provider]
	if !ok {
		return true
	}

	return time.Since(fetchedAt) > cm.cacheDuration
}


func (cm *CacheManager) CanRefresh(cache *Cache, provider string) bool {
	fetchedAt, ok := cache.FetchedAt[provider]
	if !ok {
		return true
	}

	return time.Since(fetchedAt) > cm.minRefreshInterval
}


func (cm *CacheManager) UpdateProvider(cache *Cache, provider string, models []Model) {
	
	var found bool
	for i := range cache.Catalog.Providers {
		if cache.Catalog.Providers[i].Provider.Name == provider {
			cache.Catalog.Providers[i].Models = models
			cache.Catalog.Providers[i].Updated = time.Now()
			found = true
			break
		}
	}

	
	if !found {
		if config, ok := GetProviderConfig(provider); ok {
			cache.Catalog.Providers = append(cache.Catalog.Providers, ProviderModels{
				Provider: config,
				Models:   models,
				Updated:  time.Now(),
			})
		}
	}

	
	cache.FetchedAt[provider] = time.Now()

	
	cache.Catalog.Metadata.LastUpdated = time.Now()
}


func (cm *CacheManager) Clear() error {
	if err := os.Remove(cm.cacheFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	return nil
}


func (cm *CacheManager) GetCacheAge(cache *Cache, provider string) time.Duration {
	fetchedAt, ok := cache.FetchedAt[provider]
	if !ok {
		return time.Duration(1<<63 - 1) 
	}
	return time.Since(fetchedAt)
}


func (cm *CacheManager) WarmCache() (*Cache, error) {
	cache, err := cm.Load()
	if err != nil {
		return nil, err
	}

	
	for providerName := range StaticModels {
		models := GetStaticModels(providerName)
		if len(models) > 0 {
			cm.UpdateProvider(cache, providerName, models)
		}
	}

	
	if err := cm.Save(cache); err != nil {
		return nil, err
	}

	return cache, nil
}
