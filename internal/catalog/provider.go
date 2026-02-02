package catalog

import (
	"context"
	"fmt"
)

// ModelProvider defines the interface for fetching models from different providers
type ModelProvider interface {
	// Name returns the provider identifier
	Name() string

	// Config returns the provider configuration
	Config() ProviderConfig

	// FetchModels retrieves the current list of available models from the provider API
	// Returns error if the provider doesn't support dynamic fetching or if the request fails
	FetchModels(ctx context.Context, apiKey string) ([]Model, error)

	// SupportsDynamicFetch returns true if this provider supports fetching models dynamically
	SupportsDynamicFetch() bool

	// ValidateAPIKey checks if the provided API key is valid
	ValidateAPIKey(ctx context.Context, apiKey string) error

	// GetDefaultModels returns the static fallback list of models
	GetDefaultModels() []Model
}

// ProviderFactory creates provider instances
type ProviderFactory struct {
	providers map[string]ModelProvider
}

// NewProviderFactory creates a new factory with all registered providers
func NewProviderFactory() *ProviderFactory {
	factory := &ProviderFactory{
		providers: make(map[string]ModelProvider),
	}

	// Register all built-in providers
	factory.Register(NewAnthropicProvider())
	factory.Register(NewOpenAIProvider())
	factory.Register(NewGroqProvider())
	factory.Register(NewOpenRouterProvider())
	factory.Register(NewOllamaProvider())

	return factory
}

// Register adds a provider to the factory
func (f *ProviderFactory) Register(provider ModelProvider) {
	f.providers[provider.Name()] = provider
}

// Get returns a provider by name
func (f *ProviderFactory) Get(name string) (ModelProvider, error) {
	provider, ok := f.providers[name]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
	return provider, nil
}

// List returns all registered provider names
func (f *ProviderFactory) List() []string {
	names := make([]string, 0, len(f.providers))
	for name := range f.providers {
		names = append(names, name)
	}
	return names
}

// Fetcher handles the actual HTTP requests to provider APIs
type Fetcher struct {
	client HTTPClient
}

// HTTPClient interface for making HTTP requests
// This can be implemented with the standard http.Client or a custom client
type HTTPClient interface {
	Get(ctx context.Context, url string, headers map[string]string) ([]byte, error)
}

// NewFetcher creates a new model fetcher
func NewFetcher(client HTTPClient) *Fetcher {
	if client == nil {
		client = &defaultHTTPClient{}
	}
	return &Fetcher{client: client}
}

// Fetch attempts to fetch models from a provider
func (f *Fetcher) Fetch(ctx context.Context, provider ModelProvider, apiKey string) ([]Model, error) {
	if !provider.SupportsDynamicFetch() {
		return nil, fmt.Errorf("provider %s does not support dynamic model fetching", provider.Name())
	}

	return provider.FetchModels(ctx, apiKey)
}

// defaultHTTPClient is a simple wrapper around http.Client
type defaultHTTPClient struct{}

func (c *defaultHTTPClient) Get(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	// Implementation would use http.Client with context
	// This is a placeholder - actual implementation would import net/http
	return nil, fmt.Errorf("not implemented")
}
