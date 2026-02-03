package catalog

import (
	"context"
	"fmt"
)


type ModelProvider interface {
	
	Name() string

	
	Config() ProviderConfig

	
	
	FetchModels(ctx context.Context, apiKey string) ([]Model, error)

	
	SupportsDynamicFetch() bool

	
	ValidateAPIKey(ctx context.Context, apiKey string) error

	
	GetDefaultModels() []Model
}


type ProviderFactory struct {
	providers map[string]ModelProvider
}


func NewProviderFactory() *ProviderFactory {
	factory := &ProviderFactory{
		providers: make(map[string]ModelProvider),
	}

	
	factory.Register(NewAnthropicProvider())
	factory.Register(NewOpenAIProvider())
	factory.Register(NewGroqProvider())
	factory.Register(NewOpenRouterProvider())
	factory.Register(NewOllamaProvider())

	return factory
}


func (f *ProviderFactory) Register(provider ModelProvider) {
	f.providers[provider.Name()] = provider
}


func (f *ProviderFactory) Get(name string) (ModelProvider, error) {
	provider, ok := f.providers[name]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
	return provider, nil
}


func (f *ProviderFactory) List() []string {
	names := make([]string, 0, len(f.providers))
	for name := range f.providers {
		names = append(names, name)
	}
	return names
}


type Fetcher struct {
	client HTTPClient
}



type HTTPClient interface {
	Get(ctx context.Context, url string, headers map[string]string) ([]byte, error)
}


func NewFetcher(client HTTPClient) *Fetcher {
	if client == nil {
		client = &defaultHTTPClient{}
	}
	return &Fetcher{client: client}
}


func (f *Fetcher) Fetch(ctx context.Context, provider ModelProvider, apiKey string) ([]Model, error) {
	if !provider.SupportsDynamicFetch() {
		return nil, fmt.Errorf("provider %s does not support dynamic model fetching", provider.Name())
	}

	return provider.FetchModels(ctx, apiKey)
}


type defaultHTTPClient struct{}

func (c *defaultHTTPClient) Get(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	
	
	return nil, fmt.Errorf("not implemented")
}
