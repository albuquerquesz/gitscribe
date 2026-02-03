package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)


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


func ExampleUsage() {
	
	opts := ManagerOptions{
		CacheOptions: CacheOptions{
			CacheDuration:      24 * time.Hour,
			MinRefreshInterval: 1 * time.Hour,
		},
		APIKeyResolver: func(provider string) (string, error) {
			
			return "your-api-key", nil
		},
	}

	manager, err := NewCatalogManager(opts)
	if err != nil {
		panic(err)
	}

	
	providers := manager.ListProviders()
	fmt.Println("Available providers:", providers)

	
	models, err := manager.GetModelsByProvider("openai")
	if err != nil {
		panic(err)
	}
	fmt.Printf("OpenAI models: %d\n", len(models))

	
	model, err := manager.GetModel("gpt-4o")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found model: %s\n", model.Name)

	
	filtered := manager.FilterModels(FilterOptions{
		Provider:    "openai",
		PricingTier: PricingStandard,
	})
	fmt.Printf("Filtered models: %d\n", len(filtered))

	
	recommended := manager.GetRecommendedModels("coding")
	fmt.Printf("Recommended for coding: %d\n", len(recommended))

	
	status := manager.GetCacheStatus()
	for provider, s := range status {
		fmt.Printf("%s cache: age=%v, stale=%v\n", provider, s.Age, s.IsStale)
	}

	
	ctx := context.Background()
	if err := manager.RefreshProvider(ctx, "groq"); err != nil {
		fmt.Printf("Failed to refresh: %v\n", err)
	}

	
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
