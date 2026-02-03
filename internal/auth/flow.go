package auth

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/zalando/go-keyring"
)

type Flow struct {
	config *FlowConfig
}

func NewFlow(config *FlowConfig) *Flow {
	return &Flow{
		config: config,
	}
}

func (f *Flow) Run(ctx context.Context) (*TokenResponse, string, error) {
	provider := f.config.Provider

	pkce, err := GeneratePKCE()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate PKCE: %w", err)
	}
	defer pkce.ClearVerifier()

	state, err := GenerateState()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate state: %w", err)
	}

	server, port, err := NewCallbackServer(f.config.Port)
	if err != nil {
		return nil, "", fmt.Errorf("failed to start callback server: %w", err)
	}
	defer server.Stop(context.Background())

	server.SetState(state)

	callbackPath := "/callback"
	if cp, ok := provider.(CallbackPathProvider); ok {
		callbackPath = cp.CallbackPath()
	}

	redirectURL := fmt.Sprintf("http://localhost:%d%s", port, callbackPath)

	authURL := f.buildAuthorizationURL(provider, pkce.Challenge, state, redirectURL)

	if f.config.OpenBrowser && CanOpenBrowser() {
		browser := NewBrowserOpener()
		if err := browser.Open(authURL); err != nil {
			fmt.Printf("Warning: Could not open browser automatically\n")
		}
	} else {
		fmt.Printf("\nPlease visit this URL to authorize:\n%s\n\n", authURL)
	}

	fmt.Println("Waiting for authentication...")

	callbackCtx, cancel := context.WithTimeout(ctx, f.config.Timeout)
	defer cancel()

	result, err := server.WaitForCallback(callbackCtx)
	if err != nil {
		if err == context.DeadlineExceeded {
			return nil, "", ErrTimeout
		}
		return nil, "", fmt.Errorf("failed to receive callback: %w", err)
	}

	if result.Error != "" {
		return nil, "", fmt.Errorf("OAuth error: %s", result.Error)
	}

	fmt.Println("Exchanging authorization code for tokens...")

	tokenCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tokens, err := ExchangeCode(tokenCtx, provider, result.Code, redirectURL, pkce.Verifier)
	if err != nil {
		return nil, "", err
	}

	fmt.Printf("✓ Successfully authenticated with %s\n", provider.Name())

	fmt.Println("Generating API key...")

	apiKeyCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiKey, err := provider.GenerateAPIKey(apiKeyCtx, tokens.AccessToken)
	if err != nil {
		return nil, "", err
	}

	fmt.Printf("✓ API key generated successfully\n")

	return tokens, apiKey, nil
}

func (f *Flow) buildAuthorizationURL(provider Provider, codeChallenge, state, redirectURL string) string {
	params := url.Values{
		"response_type":         {"code"},
		"client_id":             {provider.ClientID()},
		"redirect_uri":          {redirectURL},
		"scope":                 {f.buildScopeString(provider.Scopes())},
		"state":                 {state},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {"S256"},
	}

	if ep, ok := provider.(ExtraParamsProvider); ok {
		for k, v := range ep.ExtraParams() {
			params.Set(k, v)
		}
	}

	return fmt.Sprintf("%s?%s", provider.AuthorizationEndpoint(), params.Encode())
}

func (f *Flow) buildScopeString(scopes []string) string {
	result := ""
	for i, scope := range scopes {
		if i > 0 {
			result += " "
		}
		result += scope
	}
	return result
}

func IsAuthenticated(providerName string) (bool, error) {
	storage, err := NewTokenStorage()
	if err != nil {
		return false, err
	}

	token, err := storage.LoadToken(providerName)
	if err != nil {
		if err == keyring.ErrNotFound {
			return false, nil
		}
		return false, err
	}

	return !token.IsExpired(), nil
}

func RefreshIfNeeded(ctx context.Context, provider Provider) (*TokenResponse, error) {
	storage, err := NewTokenStorage()
	if err != nil {
		return nil, err
	}

	token, err := storage.LoadToken(provider.Name())
	if err != nil {
		return nil, err
	}

	if !token.NeedsRefresh() {
		return &TokenResponse{
			AccessToken:  token.AccessToken,
			TokenType:    token.TokenType,
			ExpiresAt:    token.ExpiresAt,
			RefreshToken: token.RefreshToken,
			Scope:        token.Scope,
		},
		nil
	}

	if token.RefreshToken == "" {
		return nil, fmt.Errorf("token expired and no refresh token available")
	}

	fmt.Println("Refreshing access token...")

	newToken, err := RefreshToken(ctx, provider, token.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	if err := storage.SaveToken(provider.Name(), newToken); err != nil {
		return nil, fmt.Errorf("failed to save refreshed token: %w", err)
	}

	return newToken, nil
}