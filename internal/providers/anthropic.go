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
	// Anthropic OAuth2 endpoints (these are example endpoints - replace with actual Anthropic endpoints)
	anthropicAuthEndpoint   = "https://api.anthropic.com/oauth/authorize"
	anthropicTokenEndpoint  = "https://api.anthropic.com/oauth/token"
	anthropicAPIKeyEndpoint = "https://api.anthropic.com/v1/keys"

	// Public client ID for PKCE flow (no client secret needed)
	anthropicClientID = "gitscribe-cli-public"
)

// AnthropicScopes defines the required OAuth scopes
var AnthropicScopes = []string{
	"org:create_api_key",
	"user:profile",
	"user:inference",
}

// AnthropicProvider implements the OAuth2 provider interface for Anthropic
type AnthropicProvider struct {
	baseURL string
}

// NewAnthropicProvider creates a new Anthropic OAuth provider
func NewAnthropicProvider() *AnthropicProvider {
	return &AnthropicProvider{
		baseURL: "https://api.anthropic.com",
	}
}

// NewAnthropicProviderWithBaseURL creates a provider with a custom base URL (for testing/enterprise)
func NewAnthropicProviderWithBaseURL(baseURL string) *AnthropicProvider {
	return &AnthropicProvider{
		baseURL: baseURL,
	}
}

// Name returns the provider name
func (a *AnthropicProvider) Name() string {
	return "anthropic"
}

// AuthorizationEndpoint returns the OAuth2 authorization URL
func (a *AnthropicProvider) AuthorizationEndpoint() string {
	return a.baseURL + "/oauth/authorize"
}

// TokenEndpoint returns the OAuth2 token exchange URL
func (a *AnthropicProvider) TokenEndpoint() string {
	return a.baseURL + "/oauth/token"
}

// Scopes returns the required OAuth2 scopes
func (a *AnthropicProvider) Scopes() []string {
	return AnthropicScopes
}

// ClientID returns the OAuth2 client ID
func (a *AnthropicProvider) ClientID() string {
	return anthropicClientID
}

// SupportsPKCE returns true as Anthropic supports PKCE
func (a *AnthropicProvider) SupportsPKCE() bool {
	return true
}

// APIKeyEndpoint returns the API key generation endpoint
func (a *AnthropicProvider) APIKeyEndpoint() string {
	return a.baseURL + "/v1/admin/keys"
}

// APIKeyRequest represents the request body for creating an API key
type APIKeyRequest struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes,omitempty"`
}

// APIKeyResponse represents the response from creating an API key
type APIKeyResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Key       string    `json:"key"`
	Scopes    []string  `json:"scopes"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
}

// GenerateAPIKey generates a new API key using the access token
func (a *AnthropicProvider) GenerateAPIKey(ctx context.Context, accessToken string) (string, error) {
	// Create API key request
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

	// Make request to create API key
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

// VerifyToken verifies that an access token is valid
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

// Ensure AnthropicProvider implements the Provider interface
var _ auth.Provider = (*AnthropicProvider)(nil)
