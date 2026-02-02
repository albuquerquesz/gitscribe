package providers

import (
	"context"
	"fmt"

	"github.com/albuquerquesz/gitscribe/internal/auth"
)

const (
	// OpenAI OAuth2 endpoints (Example endpoints)
	openAIAuthEndpoint  = "https://openai.com/oauth/authorize"
	openAITokenEndpoint = "https://api.openai.com/oauth/token"
)

// OpenAIScopes defines the required OAuth scopes
var OpenAIScopes = []string{
	"user.read",
	"models.read",
	"completions.write",
}

// OpenAIProvider implements the OAuth2 provider interface for OpenAI
type OpenAIProvider struct {
	baseURL string
}

// NewOpenAIProvider creates a new OpenAI OAuth provider
func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{
		baseURL: "https://api.openai.com",
	}
}

// Name returns the provider name
func (o *OpenAIProvider) Name() string {
	return "openai"
}

// AuthorizationEndpoint returns the OAuth2 authorization URL
func (o *OpenAIProvider) AuthorizationEndpoint() string {
	return openAIAuthEndpoint
}

// TokenEndpoint returns the OAuth2 token exchange URL
func (o *OpenAIProvider) TokenEndpoint() string {
	return openAITokenEndpoint
}

// Scopes returns the required OAuth2 scopes
func (o *OpenAIProvider) Scopes() []string {
	return OpenAIScopes
}

// ClientID returns the OAuth2 client ID (example)
func (o *OpenAIProvider) ClientID() string {
	return "openai-public-client"
}

// SupportsPKCE returns true
func (o *OpenAIProvider) SupportsPKCE() bool {
	return true
}

// APIKeyEndpoint returns the API key generation endpoint
func (o *OpenAIProvider) APIKeyEndpoint() string {
	return o.baseURL + "/v1/api-keys"
}

// GenerateAPIKey generates a new API key using the access token
func (o *OpenAIProvider) GenerateAPIKey(ctx context.Context, accessToken string) (string, error) {
	// Note: OpenAI might not support generating long-lived API keys via OAuth access tokens directly
	// for all users. This is a placeholder for the flow.
	return accessToken, nil
}

// Ensure OpenAIProvider implements the Provider interface
var _ auth.Provider = (*OpenAIProvider)(nil)
