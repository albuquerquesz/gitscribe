# Model Catalog System

A comprehensive system for managing AI provider models in the GitScribe CLI.

## Overview

The model catalog system provides:
- **Static model definitions**: Fallback data for all supported providers
- **Dynamic model fetching**: Real-time model lists from provider APIs (where supported)
- **Smart caching**: Local cache with configurable TTL and refresh intervals
- **Model recommendations**: Suggest models based on requirements
- **Provider abstraction**: Unified interface across all AI providers

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      CatalogManager                              │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │   Static Models │  │  Dynamic Fetch  │  │     Cache       │  │
│  │   (Fallback)    │  │   (Real-time)   │  │   (Local JSON)  │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
   ┌────▼────┐          ┌────▼────┐          ┌────▼────┐
   │ Anthropic│          │ OpenAI  │          │  Groq   │
   │ (Static) │          │ (Fetch) │          │ (Fetch) │
   └─────────┘          └─────────┘          └─────────┘
        │                     │                     │
   ┌────▼────┐          ┌────▼────┐
   │OpenRouter│          │ Ollama  │
   │ (Fetch)  │          │ (Fetch) │
   └─────────┘          └─────────┘
```

## Data Structures

### Model
```go
type Model struct {
    ID              string       // Unique identifier
    Provider        string       // Provider name
    Name            string       // Display name
    Description     string       // Human-readable description
    MaxTokens       int          // Maximum output tokens
    InputPrice      float64      // Price per 1M input tokens
    OutputPrice     float64      // Price per 1M output tokens
    PricingTier     PricingTier  // free/budget/standard/premium/enterprise
    Capabilities    []Capability // chat/vision/code/reasoning/etc
    Status          ModelStatus  // available/preview/beta/legacy
    ContextWindow   int          // Maximum context window
    TrainingCutoff  string       // Knowledge cutoff date
    SupportsVision  bool         // Can process images
    SupportsToolUse bool         // Supports function calling
    RecommendedFor  []string     // Use case recommendations
    Aliases         []string     // Alternative names
}
```

### ProviderConfig
```go
type ProviderConfig struct {
    Name           string            // Provider identifier
    BaseURL        string            // API base URL
    AuthMethod     AuthMethod        // api_key/bearer/basic/none
    DefaultHeaders map[string]string // Default request headers
    ModelsEndpoint string            // Endpoint for listing models
    SupportsList   bool              // Supports dynamic fetching
    RequiresAuth   bool              // Requires authentication
    RateLimitRPS   int               // Requests per second limit
}
```

## Providers

| Provider | Dynamic Fetch | Auth Method | Base URL |
|----------|--------------|-------------|----------|
| Anthropic | No | API Key | `https://api.anthropic.com/v1` |
| OpenAI | Yes | Bearer Token | `https://api.openai.com/v1` |
| Groq | Yes | Bearer Token | `https://api.groq.com/openai/v1` |
| OpenRouter | Yes | Bearer Token | `https://openrouter.ai/api/v1` |
| Ollama | Yes | None | `http://localhost:11434/v1` |

## Usage Examples

### Initialize the Catalog Manager

```go
import "github.com/albuquerquesz/gitscribe/internal/catalog"

// Create manager with cache and API key resolution
opts := catalog.ManagerOptions{
    CacheOptions: catalog.CacheOptions{
        CacheDuration:      24 * time.Hour,  // Cache for 24 hours
        MinRefreshInterval: 1 * time.Hour,   // Max 1 refresh per hour
    },
    APIKeyResolver: func(provider string) (string, error) {
        // Fetch from keyring or secure storage
        return store.GetProviderKey(provider)
    },
}

manager, err := catalog.NewCatalogManager(opts)
if err != nil {
    log.Fatal(err)
}
```

### List All Models for a Provider

```go
models, err := manager.GetModelsByProvider("groq")
if err != nil {
    log.Fatal(err)
}

for _, model := range models {
    fmt.Printf("- %s (%s): %s\n", model.Name, model.ID, model.PricingTier)
}
```

### Get a Specific Model

```go
model, err := manager.GetModel("llama-3.3-70b-versatile")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Model: %s\n", model.Name)
fmt.Printf("Context: %d tokens\n", model.ContextWindow)
fmt.Printf("Price: $%.2f/1M tokens\n", model.InputPrice)
```

### Filter Models

```go
// Find all budget models with code capability
filtered := manager.FilterModels(catalog.FilterOptions{
    PricingTier:    catalog.PricingBudget,
    Capability:     catalog.CapabilityCode,
    MinContextSize: 32000,
})

for _, model := range filtered {
    fmt.Println(model.String())
}
```

### Refresh Provider (Dynamic Fetch)

```go
ctx := context.Background()

// Refresh a specific provider
err := manager.RefreshProvider(ctx, "groq")
if err != nil {
    log.Printf("Failed to refresh: %v\n", err)
}

// Refresh all providers that support dynamic fetching
err = manager.RefreshAll(ctx)
```

### Get Model Recommendations

```go
// Get models recommended for coding
recommended := manager.GetRecommendedModels("coding")

// Get a specific recommendation based on requirements
suggested, err := manager.SuggestModel(catalog.ModelRequirements{
    Provider:       "groq",
    MinContextSize: 128000,
    MaxPrice:       1.0,
    RequiresVision: false,
    Capabilities:   []catalog.Capability{
        catalog.CapabilityChat,
        catalog.CapabilityCode,
    },
})
```

### Check Cache Status

```go
status := manager.GetCacheStatus()
for provider, s := range status {
    fmt.Printf("%s:\n", provider)
    fmt.Printf("  Age: %v\n", s.Age)
    fmt.Printf("  Stale: %v\n", s.IsStale)
    fmt.Printf("  Can Refresh: %v\n", s.CanRefresh)
}
```

## Cache Strategy

The caching system balances freshness with API rate limits:

1. **Cache Duration**: Default 24 hours
   - After this period, data is considered "stale"
   - Automatic refresh on next access (if API key available)

2. **Minimum Refresh Interval**: Default 1 hour
   - Prevents excessive API calls
   - Manual refresh bypasses this limit

3. **Cache Location**: `~/.multiagent/cache/model-catalog-cache.json`

4. **Warm Cache**: Static models pre-populated on first run

## Model Metadata

### Essential Fields
- `id`: Unique identifier used in API calls
- `provider`: Provider name
- `max_tokens`: Maximum tokens in output
- `context_window`: Maximum input context
- `capabilities`: What the model can do

### Nice-to-Have Fields
- `input_price`/`output_price`: Pricing information
- `training_cutoff`: Knowledge freshness
- `recommended_for`: Use case guidance
- `aliases`: Alternative names users might use

### Capability Flags
- `supports_vision`: Can process images
- `supports_tool_use`: Supports function calling
- `status`: Availability status

## Updating the Catalog

### Automatic Updates
The catalog automatically refreshes when:
1. Cache is stale (>24 hours old)
2. User explicitly requests refresh
3. First initialization (uses static models)

### Manual Updates
```go
// Force refresh (bypasses rate limiting)
err := manager.ForceRefresh(ctx, "groq")

// Clear all cached data
err := manager.ClearCache()
```

## Special Permissions

Some models require special access:
- **OpenAI o1 models**: Beta access required
- **Anthropic Opus**: Standard API access
- **Enterprise models**: Organization-level permissions

The system handles this by:
1. Marking restricted models with appropriate status
2. Validating API keys before attempting refresh
3. Gracefully falling back to static data on permission errors

## Versioning

- **Catalog Schema Version**: `1.0`
- **Cache Format Version**: `1.0`
- **Compatibility**: Backward compatible for minor versions

## Files

- `models.go`: Core data structures
- `provider.go`: Provider interface and factory
- `providers.go`: Provider implementations
- `static.go`: Static model definitions
- `cache.go`: Caching layer
- `manager.go`: Main catalog manager
- `loader.go`: File I/O utilities
- `catalog.yaml`: Static catalog data

## Integration with CLI

```go
// In your command handler
catalogCmd := &cobra.Command{
    Use:   "catalog",
    Short: "Manage AI model catalog",
}

// Add subcommands
catalogCmd.AddCommand(&cobra.Command{
    Use:   "list [provider]",
    Short: "List available models",
    RunE: func(cmd *cobra.Command, args []string) error {
        manager, _ := catalog.NewCatalogManager(opts)
        
        if len(args) > 0 {
            models, _ := manager.GetModelsByProvider(args[0])
            // Print models
        } else {
            providers := manager.ListProviders()
            // Print providers
        }
        return nil
    },
})
```
