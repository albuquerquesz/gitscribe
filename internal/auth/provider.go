// Package auth provides OAuth2 PKCE authentication for AI providers
package auth

import (
	"context"
	"fmt"
	"time"
)

// Provider defines the interface for OAuth2 providers
type Provider interface {
	// Name returns the provider name
	Name() string

	// AuthorizationEndpoint returns the OAuth2 authorization URL
	AuthorizationEndpoint() string

	// TokenEndpoint returns the OAuth2 token exchange URL
	TokenEndpoint() string

	// Scopes returns the required OAuth2 scopes
	Scopes() []string

	// ClientID returns the OAuth2 client ID (public client for PKCE)
	ClientID() string

	// SupportsPKCE returns true if the provider supports PKCE
	SupportsPKCE() bool

	// APIKeyEndpoint returns the endpoint for generating API keys
	APIKeyEndpoint() string

	// GenerateAPIKey generates an API key using the access token
	GenerateAPIKey(ctx context.Context, accessToken string) (string, error)
}

// TokenResponse represents the OAuth2 token response
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Scope        string    `json:"scope,omitempty"`
	ExpiresAt    time.Time `json:"-"`
}

// FlowConfig contains configuration for the OAuth flow
type FlowConfig struct {
	Provider     Provider
	RedirectURL  string
	Port         int
	Timeout      time.Duration
	StateTimeout time.Duration
	OpenBrowser  bool
}

// DefaultFlowConfig returns default configuration
func DefaultFlowConfig(provider Provider) *FlowConfig {
	return &FlowConfig{
		Provider:     provider,
		RedirectURL:  fmt.Sprintf("http://localhost:%d/callback", DefaultPort),
		Port:         DefaultPort,
		Timeout:      5 * time.Minute,
		StateTimeout: 10 * time.Minute,
		OpenBrowser:  true,
	}
}

// DefaultPort is the default port for the local callback server
const DefaultPort = 8085

// AlternativePorts are fallback ports if the default is unavailable
var AlternativePorts = []int{8086, 8087, 8088, 8089, 8090}

// Errors
var (
	ErrTimeout          = fmt.Errorf("authentication timeout")
	ErrInvalidState     = fmt.Errorf("invalid state parameter")
	ErrPortInUse        = fmt.Errorf("port already in use")
	ErrBrowserOpen      = fmt.Errorf("failed to open browser")
	ErrTokenExchange    = fmt.Errorf("token exchange failed")
	ErrAPIKeyGeneration = fmt.Errorf("API key generation failed")
)
