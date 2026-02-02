package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ExchangeCode exchanges the authorization code for tokens
func ExchangeCode(ctx context.Context, provider Provider, code, redirectURL, codeVerifier string) (*TokenResponse, error) {
	tokenURL := provider.TokenEndpoint()

	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURL},
		"client_id":     {provider.ClientID()},
		"code_verifier": {codeVerifier},
	}

	var lastErr error

	// Retry logic with exponential backoff (max 3 retries)
	const maxRetries = 3
	var baseDelay = 500 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 500ms, 1s, 2s
			delay := baseDelay * time.Duration(1<<(attempt-1))
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
		if err != nil {
			return nil, fmt.Errorf("failed to create token request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")

		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("token request failed: %w", err)
			continue // Retry on network errors
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			// Check if it's a server error (5xx) - these are retryable
			if resp.StatusCode >= 500 && resp.StatusCode < 600 {
				lastErr = fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
				continue
			}
			// Client errors (4xx) are not retryable
			return nil, fmt.Errorf("%w: status %d: %s", ErrTokenExchange, resp.StatusCode, string(body))
		}

		var tr TokenResponse
		if err := json.Unmarshal(body, &tr); err != nil {
			return nil, fmt.Errorf("%w: failed to parse response: %v", ErrTokenExchange, err)
		}

		// Calculate expiration time
		if tr.ExpiresIn > 0 {
			tr.ExpiresAt = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
		}

		return &tr, nil
	}

	return nil, fmt.Errorf("%w: %v", ErrTokenExchange, lastErr)
}

// RefreshToken refreshes an access token using a refresh token
func RefreshToken(ctx context.Context, provider Provider, refreshToken string) (*TokenResponse, error) {
	tokenURL := provider.TokenEndpoint()

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {provider.ClientID()},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed (%d): %s", resp.StatusCode, string(body))
	}

	var tr TokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tr.ExpiresIn > 0 {
		tr.ExpiresAt = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	}

	// Keep the refresh token if a new one wasn't provided
	if tr.RefreshToken == "" {
		tr.RefreshToken = refreshToken
	}

	return &tr, nil
}
