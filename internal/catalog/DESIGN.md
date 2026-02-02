# Model Catalog System Design

## Executive Summary

A comprehensive model catalog system for the GitScribe CLI that manages AI provider models with static fallback definitions and dynamic API fetching capabilities. The system provides intelligent caching, model recommendations, and a unified interface across all supported providers.

## Design Goals

1. **Reliability**: Always have model data available (static fallback)
2. **Freshness**: Dynamic updates from provider APIs when possible
3. **Performance**: Smart caching to minimize API calls
4. **Usability**: Rich metadata for model selection
5. **Extensibility**: Easy to add new providers

## Architecture

### Core Components

```
┌─────────────────────────────────────────────────────────────┐
│                    CatalogManager                            │
│  (Central orchestrator - thread-safe with RWMutex)           │
└──────────────────┬──────────────────────────────────────────┘
                   │
    ┌──────────────┼──────────────┐
    │              │              │
┌───▼────┐   ┌───▼────┐   ┌───▼────┐
│Provider│   │ Cache  │   │ Static │
│Factory │   │Manager │   │Catalog │
└───┬────┘   └───┬────┘   └───┬────┘
    │            │            │
┌───▼────┐   ┌───▼────┐   ┌───▼────┐
│Provider│   │  JSON  │   │  Go    │
│Impls   │   │  File  │   │  Code  │
└────────┘   └────────┘   └────────┘
```

## Data Structures

### Primary Types

#### Model
```go
type Model struct {
    // Identification
    ID       string // API identifier (e.g., "gpt-4o")
    Provider string // Provider name
    Name     string // Display name
    
    // Capabilities
    Capabilities    []Capability // chat, vision, code, reasoning, etc.
    SupportsVision  bool
    SupportsToolUse bool
    
    // Technical Specs
    MaxTokens     int    // Max output tokens
    ContextWindow int    // Max input context
    Status        ModelStatus // available, preview, beta, legacy
    
    // Pricing
    InputPrice  float64 // Per 1M tokens
    OutputPrice float64 // Per 1M tokens
    PricingTier PricingTier // free, budget, standard, premium
    
    // Metadata
    Description    string   // Human-readable
    TrainingCutoff string   // Knowledge date
    RecommendedFor []string // Use cases
    Aliases        []string // Alternative IDs
    CreatedAt      int64    // From API
}
```

#### Provider Interface
```go
type ModelProvider interface {
    Name() string
    Config() ProviderConfig
    FetchModels(ctx context.Context, apiKey string) ([]Model, error)
    SupportsDynamicFetch() bool
    ValidateAPIKey(ctx context.Context, apiKey string) error
    GetDefaultModels() []Model
}
```

## Provider Implementations

### Anthropic
- **Dynamic Fetch**: No (no public models endpoint)
- **Fallback**: Complete static catalog
- **Models**: Claude 3.5 Sonnet, Haiku, Opus
- **Update Strategy**: Manual updates when new models released

### OpenAI
- **Dynamic Fetch**: Yes
- **Endpoint**: GET /v1/models
- **Rate Limit**: 100 RPS
- **Fallback**: Static catalog for metadata enrichment
- **Special Handling**: 
  - Filter to chat models only
  - Enrich with static pricing/context data

### Groq
- **Dynamic Fetch**: Yes
- **Endpoint**: GET /openai/v1/models
- **Rate Limit**: 30 RPS
- **Fallback**: Static catalog
- **Special Handling**:
  - Provides context_window in response
  - Match static data for capabilities

### OpenRouter
- **Dynamic Fetch**: Yes
- **Endpoint**: GET /api/v1/models
- **Rate Limit**: 20 RPS
- **Fallback**: Static catalog
- **Special Handling**:
  - Rich pricing data from API
  - Requires HTTP-Referer and X-Title headers
  - Calculate pricing tier dynamically

### Ollama
- **Dynamic Fetch**: Yes (with fallback)
- **Endpoints**: 
  - /v1/models (OpenAI compatible)
  - /api/tags (Native Ollama API)
- **Rate Limit**: None (local)
- **Fallback**: Static catalog with common models
- **Special Handling**:
  - Try OpenAI endpoint first
  - Fallback to native /api/tags
  - Parse model tags (e.g., ":latest")

## Caching Strategy

### Cache Configuration
```go
type CacheOptions struct {
    CacheDuration      time.Duration // Default: 24 hours
    MinRefreshInterval time.Duration // Default: 1 hour
    CacheDir           string        // Default: ~/.multiagent/cache
}
```

### Cache Behavior

1. **Warm Cache**: On first run, populate with static models
2. **Stale Detection**: Cache >24 hours old considered stale
3. **Rate Limiting**: Max 1 refresh per hour per provider
4. **Graceful Degradation**: Use static data if fetch fails

### Cache File Format
```json
{
  "catalog": {
    "metadata": {
      "version": "1.0",
      "last_updated": "2024-01-15T10:30:00Z"
    },
    "providers": [...]
  },
  "fetched_at": {
    "groq": "2024-01-15T10:30:00Z",
    "openai": "2024-01-14T08:15:00Z"
  }
}
```

## Update Mechanisms

### Automatic Updates
- Triggered when accessing stale data
- Respects min refresh interval
- Silent fallback to static on error

### Manual Updates
```bash
# Refresh specific provider
gitscribe catalog refresh groq

# Force refresh (bypass rate limits)
gitscribe catalog refresh groq --force

# Refresh all providers
gitscribe catalog refresh
```

### Update Sources

| Priority | Source | When Used |
|----------|--------|-----------|
| 1 | Provider API | Dynamic fetch succeeds |
| 2 | Static Catalog | Dynamic fetch fails or unsupported |
| 3 | Cached Data | Within cache duration |

## Model Selection & Recommendations

### Use Case Recommendations
Each model has `RecommendedFor` tags:
- `coding`: Optimized for code generation
- `analysis`: Good for data analysis
- `reasoning`: Strong logical reasoning
- `chat`: Conversational tasks
- `fast`: Quick responses
- `cost-effective`: Low price
- `long-context`: Large context window

### Smart Suggestion Algorithm
```go
type ModelRequirements struct {
    Provider       string
    MinContextSize int
    MaxPrice       float64
    RequiresVision bool
    RequiresTools  bool
    Capabilities   []Capability
}

// Scoring:
// - Provider match: +50
// - Context size met: +20
// - Price within budget: +10 * savings
// - Capabilities: +10 each
// - Vision/tools bonus: +5 each
// - Recency bonus: +days since epoch
```

## Security Considerations

### API Key Handling
- Never store API keys in catalog files
- Use keyring for secure storage
- API key resolver passed at initialization
- Keys only used for fetch operations

### Cache Security
- Cache file permissions: 0600 (user only)
- No API keys in cache
- Public model metadata only

## Metadata Fields Priority

### Essential (Required)
- `id`: API identifier
- `provider`: Provider name
- `max_tokens`: Output limit
- `context_window`: Input limit
- `capabilities`: Core capabilities
- `status`: Availability

### Important (Recommended)
- `name`: Display name
- `pricing_tier`: Cost category
- `supports_vision`: Image processing
- `supports_tool_use`: Function calling

### Nice-to-Have (Optional)
- `description`: Human-readable
- `input_price`/`output_price`: Exact pricing
- `training_cutoff`: Knowledge date
- `recommended_for`: Use cases
- `aliases`: Alternative names

## Versioning

### Schema Versioning
- Catalog format: `1.0`
- Cache format: `1.0`
- Backward compatible for minor versions
- Migration path for major versions

### Model Updates
- Static catalog updated manually
- Dynamic fetch pulls latest from APIs
- Breaking changes handled via `status` field

## Files Structure

```
internal/catalog/
├── models.go       # Core data structures
├── provider.go     # Provider interface & factory
├── providers.go    # Provider implementations
├── static.go       # Static model definitions
├── cache.go        # Caching layer
├── manager.go      # Main catalog manager
├── loader.go       # File I/O utilities
├── catalog.yaml    # Static catalog data
└── README.md       # Documentation
```

## CLI Integration

### Commands
```bash
# List providers and models
gitscribe catalog list
gitscribe catalog list groq --details
gitscribe catalog list --tier=budget --capability=code

# Show model details
gitscribe catalog show gpt-4o
gitscribe catalog show llama-3.3-70b-versatile

# Refresh from APIs
gitscribe catalog refresh
gitscribe catalog refresh groq --force

# Check cache status
gitscribe catalog status

# Get recommendations
gitscribe catalog suggest --use-case=coding
gitscribe catalog suggest --provider=groq --max-price=1.0 --tools
```

## Questions & Answers

### Should we cache model lists locally?
**Yes**, with the following rationale:
- Reduces API calls and improves performance
- Allows offline operation with static fallback
- 24-hour TTL balances freshness with efficiency
- Cache is user-specific (no shared state issues)

### How often to refresh from APIs?
**Default: 24 hours**
- Most providers don't change models frequently
- Rate limiting considerations (20-100 RPS)
- Manual refresh available for immediate updates
- Min refresh interval: 1 hour (configurable)

### What model metadata is essential vs nice-to-have?
**Essential**: ID, provider, max_tokens, context_window, capabilities, status
**Important**: Name, pricing_tier, supports_vision, supports_tool_use
**Nice-to-have**: Description, exact pricing, training cutoff, recommendations, aliases

### How to handle models that require special permissions?
- Mark with `status: preview` or `status: beta`
- Validate API key permissions during refresh
- Gracefully skip models that return 403 Forbidden
- Document requirements in model description

## Future Enhancements

1. **Streaming Updates**: WebSocket/SSE for real-time model changes
2. **Custom Providers**: User-defined provider configurations
3. **Model Comparison**: Side-by-side model comparison
4. **Usage Analytics**: Track which models are most used
5. **Performance Metrics**: Latency/quality benchmarks per model
6. **Cost Tracking**: Estimate costs based on usage patterns

## Implementation Checklist

- [x] Core data structures (Model, ProviderConfig, etc.)
- [x] Provider interface definition
- [x] Static model definitions for all providers
- [x] Provider implementations (Anthropic, OpenAI, Groq, OpenRouter, Ollama)
- [x] Caching layer with TTL and rate limiting
- [x] Catalog manager with thread-safe operations
- [x] File I/O utilities (YAML/JSON)
- [x] Static catalog YAML file
- [x] CLI commands integration
- [x] Documentation (README, design doc)
- [ ] Unit tests for all components
- [ ] Integration tests for provider APIs
- [ ] Performance benchmarks
- [ ] Update existing AI client to use catalog
