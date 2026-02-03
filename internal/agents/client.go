package agents

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/secrets"
	openai "github.com/sashabaranov/go-openai"
)

type Client interface {
	SendMessage(ctx context.Context, messages []Message, options RequestOptions) (*Response, error)
	GetProvider() config.AgentProvider
	GetModel() string
	IsAvailable() bool
	Close() error
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestOptions struct {
	Temperature float32
	MaxTokens   int
	Timeout     time.Duration
	Stream      bool
}

type Response struct {
	Content      string
	Usage        Usage
	FinishReason string
	Model        string
}

type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type OpenAIClient struct {
	client   *openai.Client
	profile  config.AgentProfile
	apiKey   string
	provider config.AgentProvider
}

func NewOpenAIClient(profile config.AgentProfile, apiKey string) (*OpenAIClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required for agent: %s", profile.Name)
	}

	cfg := openai.DefaultConfig(apiKey)

	baseURL := profile.BaseURL
	if baseURL == "" {
		switch profile.Provider {
		case config.ProviderGroq:
			baseURL = "https://api.groq.com/openai/v1"
		case config.ProviderOpenRouter:
			baseURL = "https://openrouter.ai/api/v1"
		case config.ProviderOllama:
			baseURL = "http://localhost:11434/v1"
		case config.ProviderOpenCode:
			baseURL = "https://api.opencode.com/v1"
		default:
			baseURL = "https://api.openai.com/v1"
		}
	}

	cfg.BaseURL = strings.TrimSuffix(baseURL, "/")
	client := openai.NewClientWithConfig(cfg)

	return &OpenAIClient{
		client:   client,
		profile:  profile,
		apiKey:   apiKey,
		provider: profile.Provider,
	}, nil
}

func (c *OpenAIClient) SendMessage(ctx context.Context, messages []Message, options RequestOptions) (*Response, error) {
	if options.Timeout == 0 {
		options.Timeout = time.Duration(c.profile.Timeout) * time.Second
	}
	if options.Timeout == 0 {
		options.Timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()

	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	if c.profile.SystemPrompt != "" && len(openaiMessages) > 0 && openaiMessages[0].Role != "system" {
		openaiMessages = append([]openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: c.profile.SystemPrompt,
			},
		}, openaiMessages...)
	}

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

func (c *OpenAIClient) GetProvider() config.AgentProvider {
	return c.provider
}

func (c *OpenAIClient) GetModel() string {
	return c.profile.Model
}

func (c *OpenAIClient) IsAvailable() bool {
	return c.client != nil && c.apiKey != ""
}

func (c *OpenAIClient) Close() error {
	secrets.SecureWipe(&c.apiKey)
	return nil
}

type Factory struct {
	secretsManager *secrets.AgentKeyManager
}

func NewFactory() *Factory {
	return &Factory{
		secretsManager: secrets.NewAgentKeyManager(),
	}
}

func (f *Factory) CreateClient(profile config.AgentProfile) (Client, error) {
	apiKey, source := f.resolveAPIKey(profile)
	if apiKey == "" {
		return nil, fmt.Errorf("no API key found for agent %s (provider: %s). Configure with 'gs auth set-key -p %s' or set %s environment variable",
			profile.Name, profile.Provider, profile.Provider, getEnvKeyForProvider(profile.Provider))
	}
	_ = source // Can be used for logging/debugging

	switch profile.Provider {
	case config.ProviderOpenAI, config.ProviderGroq, config.ProviderOpenRouter, config.ProviderOllama, config.ProviderOpenCode:
		return NewOpenAIClient(profile, apiKey)
	case config.ProviderClaude:
		return NewAnthropicClient(profile, apiKey)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", profile.Provider)
	}
}

func getEnvKeyForProvider(provider config.AgentProvider) string {
	envVars := map[config.AgentProvider]string{
		config.ProviderOpenAI:     "OPENAI_API_KEY",
		config.ProviderClaude:     "ANTHROPIC_API_KEY",
		config.ProviderGroq:       "GROQ_API_KEY",
		config.ProviderOpenCode:   "OPENCODE_API_KEY",
		config.ProviderGemini:     "GOOGLE_API_KEY",
		config.ProviderOpenRouter: "OPENROUTER_API_KEY",
	}
	if envVar, ok := envVars[provider]; ok {
		return os.Getenv(envVar)
	}
	return ""
}

func (f *Factory) resolveAPIKey(profile config.AgentProfile) (apiKey string, source string) {
	if key, err := f.secretsManager.RetrieveAgentKey(profile.Name); err == nil && key != "" {
		return key, "keyring"
	}

	if key, err := f.secretsManager.Retrieve(profile.KeyringKey); err == nil && key != "" {
		return key, "keyring"
	}

	if key := getEnvKeyForProvider(profile.Provider); key != "" {
		return key, "environment"
	}

	if auth, err := secrets.LoadOpenCodeAuth(); err == nil && auth != nil {
		providerName := string(profile.Provider)
		if key, ok := auth.GetAPIKey(providerName); ok {
			return key, "opencode"
		}
	}

	return "", ""
}

func (f *Factory) CreateClientWithKey(profile config.AgentProfile, apiKey string) (Client, error) {
	switch profile.Provider {
	case config.ProviderOpenAI, config.ProviderGroq, config.ProviderOpenRouter, config.ProviderOllama, config.ProviderOpenCode:
		return NewOpenAIClient(profile, apiKey)
	case config.ProviderClaude:
		return NewAnthropicClient(profile, apiKey)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", profile.Provider)
	}
}
