package agents

import (
	"context"
	"fmt"
	"time"

	"github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/secrets"
	openai "github.com/sashabaranov/go-openai"
)

// Client defines the interface for AI clients
type Client interface {
	SendMessage(ctx context.Context, messages []Message, options RequestOptions) (*Response, error)
	GetProvider() config.AgentProvider
	GetModel() string
	IsAvailable() bool
	Close() error
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// RequestOptions contains options for a request
type RequestOptions struct {
	Temperature float32
	MaxTokens   int
	Timeout     time.Duration
	Stream      bool
}

// Response contains the AI response
type Response struct {
	Content      string
	Usage        Usage
	FinishReason string
	Model        string
}

// Usage contains token usage information
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// OpenAIClient implements Client for OpenAI-compatible APIs
type OpenAIClient struct {
	client   *openai.Client
	profile  config.AgentProfile
	apiKey   string
	provider config.AgentProvider
}

// NewOpenAIClient creates a new OpenAI-compatible client
func NewOpenAIClient(profile config.AgentProfile, apiKey string) (*OpenAIClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required for agent: %s", profile.Name)
	}

	cfg := openai.DefaultConfig(apiKey)

	// Use custom base URL if provided
	if profile.BaseURL != "" {
		cfg.BaseURL = profile.BaseURL
	}

	// Set provider-specific defaults
	switch profile.Provider {
	case config.ProviderGroq:
		if cfg.BaseURL == "" {
			cfg.BaseURL = "https://api.groq.com/openai/v1"
		}
	case config.ProviderOpenRouter:
		if cfg.BaseURL == "" {
			cfg.BaseURL = "https://openrouter.ai/api/v1"
		}
	case config.ProviderOllama:
		if cfg.BaseURL == "" {
			cfg.BaseURL = "http://localhost:11434/v1"
		}
	}

	client := openai.NewClientWithConfig(cfg)

	return &OpenAIClient{
		client:   client,
		profile:  profile,
		apiKey:   apiKey,
		provider: profile.Provider,
	}, nil
}

// SendMessage sends a message to the AI
func (c *OpenAIClient) SendMessage(ctx context.Context, messages []Message, options RequestOptions) (*Response, error) {
	// Set default timeout if not specified
	if options.Timeout == 0 {
		options.Timeout = time.Duration(c.profile.Timeout) * time.Second
	}
	if options.Timeout == 0 {
		options.Timeout = 30 * time.Second
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()

	// Convert messages to OpenAI format
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Add system prompt if configured
	if c.profile.SystemPrompt != "" && len(openaiMessages) > 0 && openaiMessages[0].Role != "system" {
		openaiMessages = append([]openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: c.profile.SystemPrompt,
			},
		}, openaiMessages...)
	}

	// Set defaults from profile if not overridden
	temperature := options.Temperature
	if temperature == 0 && c.profile.Temperature != 0 {
		temperature = c.profile.Temperature
	}
	if temperature == 0 {
		temperature = 0.7
	}

	maxTokens := options.MaxTokens
	if maxTokens == 0 && c.profile.MaxTokens != 0 {
		maxTokens = c.profile.MaxTokens
	}

	req := openai.ChatCompletionRequest{
		Model:       c.profile.Model,
		Messages:    openaiMessages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned")
	}

	return &Response{
		Content: resp.Choices[0].Message.Content,
		Usage: Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		FinishReason: string(resp.Choices[0].FinishReason),
		Model:        resp.Model,
	}, nil
}

// GetProvider returns the provider name
func (c *OpenAIClient) GetProvider() config.AgentProvider {
	return c.provider
}

// GetModel returns the model name
func (c *OpenAIClient) GetModel() string {
	return c.profile.Model
}

// IsAvailable checks if the client is properly configured
func (c *OpenAIClient) IsAvailable() bool {
	return c.client != nil && c.apiKey != ""
}

// Close cleans up resources
func (c *OpenAIClient) Close() error {
	// Clear sensitive data
	secrets.SecureWipe(&c.apiKey)
	return nil
}

// Factory creates clients based on profile
type Factory struct {
	secretsManager *secrets.AgentKeyManager
}

// NewFactory creates a new agent factory
func NewFactory() *Factory {
	return &Factory{
		secretsManager: secrets.NewAgentKeyManager(),
	}
}

// CreateClient creates a client for the given profile
func (f *Factory) CreateClient(profile config.AgentProfile) (Client, error) {
	// Retrieve API key from keyring
	apiKey, err := f.secretsManager.RetrieveAgentKey(profile.Name)
	if err != nil {
		// Fallback to the old keyring key format for backward compatibility
		apiKey, err = f.secretsManager.Retrieve(profile.KeyringKey)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve API key for agent %s: %w", profile.Name, err)
		}
	}

	switch profile.Provider {
	case config.ProviderOpenAI, config.ProviderGroq, config.ProviderOpenRouter, config.ProviderOllama:
		return NewOpenAIClient(profile, apiKey)

	// Future providers can be added here:
	// case config.ProviderClaude:
	//     return NewClaudeClient(profile, apiKey)
	// case config.ProviderGemini:
	//     return NewGeminiClient(profile, apiKey)

	default:
		return nil, fmt.Errorf("unsupported provider: %s", profile.Provider)
	}
}

// CreateClientWithKey creates a client with an explicit API key
func (f *Factory) CreateClientWithKey(profile config.AgentProfile, apiKey string) (Client, error) {
	switch profile.Provider {
	case config.ProviderOpenAI, config.ProviderGroq, config.ProviderOpenRouter, config.ProviderOllama:
		return NewOpenAIClient(profile, apiKey)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", profile.Provider)
	}
}
