
package auth

import (
	"context"
	"fmt"
	"time"
)


type Provider interface {
	
	Name() string

	
	AuthorizationEndpoint() string

	
	TokenEndpoint() string

	
	Scopes() []string

	
	ClientID() string

	
	SupportsPKCE() bool

	
	APIKeyEndpoint() string

	
	GenerateAPIKey(ctx context.Context, accessToken string) (string, error)
}


type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Scope        string    `json:"scope,omitempty"`
	ExpiresAt    time.Time `json:"-"`
}


type FlowConfig struct {
	Provider     Provider
	RedirectURL  string
	Port         int
	Timeout      time.Duration
	StateTimeout time.Duration
	OpenBrowser  bool
}


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


const DefaultPort = 8085


var AlternativePorts = []int{8086, 8087, 8088, 8089, 8090}


var (
	ErrTimeout          = fmt.Errorf("authentication timeout")
	ErrInvalidState     = fmt.Errorf("invalid state parameter")
	ErrPortInUse        = fmt.Errorf("port already in use")
	ErrBrowserOpen      = fmt.Errorf("failed to open browser")
	ErrTokenExchange    = fmt.Errorf("token exchange failed")
	ErrAPIKeyGeneration = fmt.Errorf("API key generation failed")
)
