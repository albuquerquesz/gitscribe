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
	
	anthropicAuthEndpoint   = "https://api.anthropic.com/oauth/authorize"
	anthropicTokenEndpoint  = "https://api.anthropic.com/oauth/token"
	anthropicAPIKeyEndpoint = "https://api.anthropic.com/v1/keys"

	
	anthropicClientID = "gitscribe-cli-public"
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


func NewAnthropicProviderWithBaseURL(baseURL string) *AnthropicProvider {
	return &AnthropicProvider{
		baseURL: baseURL,
	}
}


func (a *AnthropicProvider) Name() string {
	return "anthropic"
}


func (a *AnthropicProvider) AuthorizationEndpoint() string {
	return a.baseURL + "/oauth/authorize"
}


func (a *AnthropicProvider) TokenEndpoint() string {
	return a.baseURL + "/oauth/token"
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


func (a *AnthropicProvider) VerifyToken(ctx context.Context, accessToken string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+"/v1/me", nil)
	if err != nil {
		return fmt.Errorf("failed to create verify request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Anthropic-Version", "2023-06-01")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("token verification failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token invalid (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}


var _ auth.Provider = (*AnthropicProvider)(nil)
