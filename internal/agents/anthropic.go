package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/secrets"
)

const (
	defaultAnthropicBaseURL = "https://api.anthropic.com/v1"
	anthropicVersion        = "2023-06-01"
)

type AnthropicClient struct {
	client   *http.Client
	profile  config.AgentProfile
	apiKey   string
	baseURL  string
}

func NewAnthropicClient(profile config.AgentProfile, apiKey string) (*AnthropicClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required for agent: %s", profile.Name)
	}

	baseURL := defaultAnthropicBaseURL
	if profile.BaseURL != "" {
		baseURL = profile.BaseURL
	}

	return &AnthropicClient{
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		profile: profile,
		apiKey:  apiKey,
		baseURL: baseURL,
	}, nil
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	Messages  []anthropicMessage `json:"messages"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Stream    bool               `json:"stream,omitempty"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type anthropicResponse struct {
	ID           string             `json:"id"`
	Type         string             `json:"type"`
	Role         string             `json:"role"`
	Content      []anthropicContent `json:"content"`
	Usage        anthropicUsage     `json:"usage"`
	StopReason   string             `json:"stop_reason,omitempty"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func (c *AnthropicClient) SendMessage(ctx context.Context, messages []Message, options RequestOptions) (*Response, error) {
	if options.Timeout == 0 {
		options.Timeout = time.Duration(c.profile.Timeout) * time.Second
	}
	if options.Timeout == 0 {
		options.Timeout = 60 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()

	var anthropicMessages []anthropicMessage
	var systemPrompt string

	if c.profile.SystemPrompt != "" {
		systemPrompt = c.profile.SystemPrompt
	}

	for _, msg := range messages {
		if msg.Role == "system" {
			if systemPrompt == "" {
				systemPrompt = msg.Content
			}
			continue
		}
		anthropicMessages = append(anthropicMessages, anthropicMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	maxTokens := options.MaxTokens
	if maxTokens == 0 && c.profile.MaxTokens != 0 {
		maxTokens = c.profile.MaxTokens
	}
	if maxTokens == 0 {
		maxTokens = 4096
	}

	reqBody := anthropicRequest{
		Model:     c.profile.Model,
		Messages:  anthropicMessages,
		MaxTokens: maxTokens,
		System:    systemPrompt,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("content-type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic api error (%d): %s", resp.StatusCode, string(body))
	}

	var anthropicResp anthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	fullText := ""
	for _, content := range anthropicResp.Content {
		if content.Type == "text" {
			fullText += content.Text
		}
	}

	return &Response{
		Content: fullText,
		Usage: Usage{
			PromptTokens:     anthropicResp.Usage.InputTokens,
			CompletionTokens: anthropicResp.Usage.OutputTokens,
			TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		},
		FinishReason: anthropicResp.StopReason,
		Model:        c.profile.Model,
	}, nil
}

func (c *AnthropicClient) GetProvider() config.AgentProvider {
	return config.ProviderClaude
}

func (c *AnthropicClient) GetModel() string {
	return c.profile.Model
}

func (c *AnthropicClient) IsAvailable() bool {
	return c.apiKey != ""
}

func (c *AnthropicClient) Close() error {
	secrets.SecureWipe(&c.apiKey)
	return nil
}
