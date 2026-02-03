package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/albuquerquesz/gitscribe/internal/auth"
)

const (
	anthropicAuthEndpoint   = "https://claude.ai/oauth/authorize"
	anthropicTokenEndpoint  = "https://api.anthropic.com/oauth/token"
	anthropicAPIKeyEndpoint = "https://api.anthropic.com/v1/keys"
	anthropicClientID      = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"
)

var AnthropicScopes = []string{
	"org:create_api_key",
	"user:profile",
	"user:inference",
}

type AnthropicProvider struct {
	baseURL string
}

func NewAnthropicProvider() *AnthropicProvider {
	return &AnthropicProvider{
		baseURL: "https://api.anthropic.com",
	}
}

func (a *AnthropicProvider) Name() string {
	return "anthropic"
}

func (a *AnthropicProvider) AuthorizationEndpoint() string {
	return anthropicAuthEndpoint
}

func (a *AnthropicProvider) TokenEndpoint() string {
	return anthropicTokenEndpoint
}

func (a *AnthropicProvider) Scopes() []string {
	return AnthropicScopes
}

func (a *AnthropicProvider) ClientID() string {
	return anthropicClientID
}

func (a *AnthropicProvider) SupportsPKCE() bool {
	return true
}

func (a *AnthropicProvider) APIKeyEndpoint() string {
	return a.baseURL + "/v1/admin/keys"
}

func (a *AnthropicProvider) ExtraParams() map[string]string {
	return map[string]string{
		"code": "true",
	}
}

func (a *AnthropicProvider) GenerateAPIKey(ctx context.Context, accessToken string) (string, error) {
	reqBody := APIKeyRequest{
		Name: "gitscribe-cli-auto-generated",
		Scopes: []string{
			"message:write",
			"message:read",
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal API key request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.APIKeyEndpoint(), bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create API key request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Anthropic-Version", "2023-06-01")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API key generation request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read API key response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API key generation failed (%d): %s", resp.StatusCode, string(body))
	}

	var apiKeyResp APIKeyResponse
	if err := json.Unmarshal(body, &apiKeyResp); err != nil {
		return "", fmt.Errorf("failed to parse API key response: %w", err)
	}

	return apiKeyResp.Key, nil
}

type APIKeyRequest struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes,omitempty"`
}

type APIKeyResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Key       string    `json:"key"`
	Scopes    []string  `json:"scopes"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
}

func (a *AnthropicProvider) CallbackPath() string {

	return "/oauth/code/callback"

}



var _ auth.Provider = (*AnthropicProvider)(nil)

var _ auth.ExtraParamsProvider = (*AnthropicProvider)(nil)

var _ auth.CallbackPathProvider = (*AnthropicProvider)(nil)
