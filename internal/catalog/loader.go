package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// LoadFromYAML loads a catalog from a YAML file
func LoadFromYAML(path string) (*ModelCatalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read catalog file: %w", err)
	}

	var catalog ModelCatalog
	if err := yaml.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("failed to parse catalog file: %w", err)
	}

	return &catalog, nil
}

// SaveToYAML saves a catalog to a YAML file
func SaveToYAML(catalog *ModelCatalog, path string) error {
	data, err := yaml.Marshal(catalog)
	if err != nil {
		return fmt.Errorf("failed to marshal catalog: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write catalog file: %w", err)
	}

	return nil
}

// LoadFromJSON loads a catalog from a JSON file
func LoadFromJSON(path string) (*ModelCatalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read catalog file: %w", err)
	}

	var catalog ModelCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("failed to parse catalog file: %w", err)
	}

	return &catalog, nil
}

// SaveToJSON saves a catalog to a JSON file
func SaveToJSON(catalog *ModelCatalog, path string) error {
	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal catalog: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write catalog file: %w", err)
	}

	return nil
}

// ExampleUsage demonstrates how to use the catalog system
func ExampleUsage() {
	// 1. Create a catalog manager
	opts := ManagerOptions{
		CacheOptions: CacheOptions{
			CacheDuration:      24 * time.Hour,
			MinRefreshInterval: 1 * time.Hour,
		},
		APIKeyResolver: func(provider string) (string, error) {
			// This should fetch from your keyring/storage
			return "your-api-key", nil
		},
	}

	manager, err := NewCatalogManager(opts)
	if err != nil {
		panic(err)
	}

	// 2. List all providers
	providers := manager.ListProviders()
	fmt.Println("Available providers:", providers)

	// 3. Get models for a specific provider
	models, err := manager.GetModelsByProvider("openai")
	if err != nil {
		panic(err)
	}
	fmt.Printf("OpenAI models: %d\n", len(models))

	// 4. Get a specific model
	model, err := manager.GetModel("gpt-4o")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found model: %s\n", model.Name)

	// 5. Filter models by criteria
	filtered := manager.FilterModels(FilterOptions{
		Provider:    "openai",
		PricingTier: PricingStandard,
	})
	fmt.Printf("Filtered models: %d\n", len(filtered))

	// 6. Get recommended models for a use case
	recommended := manager.GetRecommendedModels("coding")
	fmt.Printf("Recommended for coding: %d\n", len(recommended))

	// 7. Get cache status
	status := manager.GetCacheStatus()
	for provider, s := range status {
		fmt.Printf("%s cache: age=%v, stale=%v\n", provider, s.Age, s.IsStale)
	}

	// 8. Refresh a provider (if API key is available)
	ctx := context.Background()
	if err := manager.RefreshProvider(ctx, "groq"); err != nil {
		fmt.Printf("Failed to refresh: %v\n", err)
	}

	// 9. Suggest a model based on requirements
	suggested, err := manager.SuggestModel(ModelRequirements{
		Provider:       "groq",
		RequiresVision: false,
		Capabilities:   []Capability{CapabilityChat, CapabilityCode},
		MaxPrice:       1.0,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Suggested model: %s\n", suggested.Name)
}
