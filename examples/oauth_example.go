package main

import (
	"context"
	"fmt"
	"log"

	"github.com/albuquerquesz/gitscribe/internal/auth"
	"github.com/albuquerquesz/gitscribe/internal/providers"
)

// Example: Authenticating with Anthropic using OAuth2 PKCE
func main() {
	// Create Anthropic provider
	provider := providers.NewAnthropicProvider()

	// Configure the OAuth flow
	flowConfig := &auth.FlowConfig{
		Provider:    provider,
		RedirectURL: "http://localhost:8085/callback",
		Port:        8085,
		Timeout:     0, // Use default
		OpenBrowser: true,
	}

	// Create and run the flow
	flow := auth.NewFlow(flowConfig)
	ctx := context.Background()

	// Run the complete OAuth flow
	// This will:
	// 1. Generate PKCE code verifier and challenge
	// 2. Start local HTTP server on localhost:8085
	// 3. Open browser to Anthropic's OAuth authorization page
	// 4. Wait for user to authenticate and authorize
	// 5. Receive the authorization code via callback
	// 6. Exchange code for access token
	// 7. Use access token to generate API key
	tokens, apiKey, err := flow.Run(ctx)
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	// Store tokens securely
	storage, err := auth.NewTokenStorage()
	if err != nil {
		log.Fatalf("Failed to create token storage: %v", err)
	}

	if err := storage.SaveToken(provider.Name(), tokens); err != nil {
		log.Fatalf("Failed to save tokens: %v", err)
	}

	// Store API key securely
	if err := auth.StoreAPIKey(provider.Name(), apiKey); err != nil {
		log.Fatalf("Failed to store API key: %v", err)
	}

	fmt.Printf("Successfully authenticated!\n")
	fmt.Printf("API Key: %s...%s\n", apiKey[:4], apiKey[len(apiKey)-4:])
}

// Example: Using stored tokens
func exampleUseStoredToken() {
	storage, _ := auth.NewTokenStorage()

	// Load stored token
	token, err := storage.LoadToken("anthropic")
	if err != nil {
		log.Fatalf("No stored token: %v", err)
	}

	// Check if token needs refresh
	if token.NeedsRefresh() && token.RefreshToken != "" {
		provider := providers.NewAnthropicProvider()
		ctx := context.Background()

		newToken, err := auth.RefreshToken(ctx, provider, token.RefreshToken)
		if err != nil {
			log.Fatalf("Failed to refresh token: %v", err)
		}

		// Save the new token
		storage.SaveToken("anthropic", newToken)
		token.AccessToken = newToken.AccessToken
	}

	// Use the access token
	fmt.Printf("Access Token: %s...\n", token.AccessToken[:10])
}

// Example: Refresh token flow
func exampleRefreshToken() {
	provider := providers.NewAnthropicProvider()
	ctx := context.Background()

	// This will automatically refresh if needed
	token, err := auth.RefreshIfNeeded(ctx, provider)
	if err != nil {
		log.Fatalf("Failed to get valid token: %v", err)
	}

	fmt.Printf("Token valid until: %v\n", token.ExpiresAt)
}
