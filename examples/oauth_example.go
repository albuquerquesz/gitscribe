package main

import (
	"context"
	"fmt"
	"log"

	"github.com/albuquerquesz/gitscribe/internal/auth"
	"github.com/albuquerquesz/gitscribe/internal/providers"
)


func main() {
	
	provider := providers.NewAnthropicProvider()

	
	flowConfig := &auth.FlowConfig{
		Provider:    provider,
		RedirectURL: "http://localhost:8085/callback",
		Port:        8085,
		Timeout:     0, 
		OpenBrowser: true,
	}

	
	flow := auth.NewFlow(flowConfig)
	ctx := context.Background()

	
	
	
	
	
	
	
	
	
	tokens, apiKey, err := flow.Run(ctx)
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	
	storage, err := auth.NewTokenStorage()
	if err != nil {
		log.Fatalf("Failed to create token storage: %v", err)
	}

	if err := storage.SaveToken(provider.Name(), tokens); err != nil {
		log.Fatalf("Failed to save tokens: %v", err)
	}

	
	if err := auth.StoreAPIKey(provider.Name(), apiKey); err != nil {
		log.Fatalf("Failed to store API key: %v", err)
	}

	fmt.Printf("Successfully authenticated!\n")
	fmt.Printf("API Key: %s...%s\n", apiKey[:4], apiKey[len(apiKey)-4:])
}


func exampleUseStoredToken() {
	storage, _ := auth.NewTokenStorage()

	
	token, err := storage.LoadToken("anthropic")
	if err != nil {
		log.Fatalf("No stored token: %v", err)
	}

	
	if token.NeedsRefresh() && token.RefreshToken != "" {
		provider := providers.NewAnthropicProvider()
		ctx := context.Background()

		newToken, err := auth.RefreshToken(ctx, provider, token.RefreshToken)
		if err != nil {
			log.Fatalf("Failed to refresh token: %v", err)
		}

		
		storage.SaveToken("anthropic", newToken)
		token.AccessToken = newToken.AccessToken
	}

	
	fmt.Printf("Access Token: %s...\n", token.AccessToken[:10])
}


func exampleRefreshToken() {
	provider := providers.NewAnthropicProvider()
	ctx := context.Background()

	
	token, err := auth.RefreshIfNeeded(ctx, provider)
	if err != nil {
		log.Fatalf("Failed to get valid token: %v", err)
	}

	fmt.Printf("Token valid until: %v\n", token.ExpiresAt)
}
