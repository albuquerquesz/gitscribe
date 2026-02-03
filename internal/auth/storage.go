package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/zalando/go-keyring"
)

const (
	serviceName = "gitscribe-oauth"
	keyringUser = "oauth-tokens"
)


type StoredToken struct {
	Provider     string    `json:"provider"`
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Scope        string    `json:"scope,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}


func (st *StoredToken) IsExpired() bool {
	return time.Now().After(st.ExpiresAt)
}


func (st *StoredToken) NeedsRefresh() bool {
	return time.Now().Add(5 * time.Minute).After(st.ExpiresAt)
}


type TokenStorage struct {
	configDir string
}


func NewTokenStorage() (*TokenStorage, error) {
	configDir, err := config.EnsureConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	return &TokenStorage{
		configDir: configDir,
	}, nil
}


func (ts *TokenStorage) SaveToken(providerName string, token *TokenResponse) error {
	
	accessTokenKey := fmt.Sprintf("%s-access-token", providerName)
	if err := keyring.Set(serviceName, accessTokenKey, token.AccessToken); err != nil {
		return fmt.Errorf("failed to store access token in keyring: %w", err)
	}

	
	if token.RefreshToken != "" {
		refreshTokenKey := fmt.Sprintf("%s-refresh-token", providerName)
		if err := keyring.Set(serviceName, refreshTokenKey, token.RefreshToken); err != nil {
			return fmt.Errorf("failed to store refresh token in keyring: %w", err)
		}
	}

	
	metadata := &StoredToken{
		Provider:  providerName,
		TokenType: token.TokenType,
		ExpiresAt: token.ExpiresAt,
		Scope:     token.Scope,
		UpdatedAt: time.Now(),
	}

	if err := ts.saveMetadata(metadata); err != nil {
		return fmt.Errorf("failed to save token metadata: %w", err)
	}

	return nil
}


func (ts *TokenStorage) LoadToken(providerName string) (*StoredToken, error) {
	
	metadata, err := ts.loadMetadata(providerName)
	if err != nil {
		return nil, err
	}

	
	accessTokenKey := fmt.Sprintf("%s-access-token", providerName)
	accessToken, err := keyring.Get(serviceName, accessTokenKey)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, fmt.Errorf("no stored token found for %s", providerName)
		}
		return nil, fmt.Errorf("failed to retrieve access token from keyring: %w", err)
	}

	
	refreshTokenKey := fmt.Sprintf("%s-refresh-token", providerName)
	refreshToken, err := keyring.Get(serviceName, refreshTokenKey)
	if err != nil && err != keyring.ErrNotFound {
		return nil, fmt.Errorf("failed to retrieve refresh token from keyring: %w", err)
	}

	return &StoredToken{
		Provider:     metadata.Provider,
		AccessToken:  accessToken,
		TokenType:    metadata.TokenType,
		ExpiresAt:    metadata.ExpiresAt,
		RefreshToken: refreshToken,
		Scope:        metadata.Scope,
		UpdatedAt:    metadata.UpdatedAt,
	}, nil
}


func (ts *TokenStorage) DeleteToken(providerName string) error {
	
	accessTokenKey := fmt.Sprintf("%s-access-token", providerName)
	refreshTokenKey := fmt.Sprintf("%s-refresh-token", providerName)

	keyring.Delete(serviceName, accessTokenKey)
	keyring.Delete(serviceName, refreshTokenKey)

	
	metadataFile := filepath.Join(ts.configDir, fmt.Sprintf("%s-token.json", providerName))
	os.Remove(metadataFile)

	return nil
}


func (ts *TokenStorage) saveMetadata(token *StoredToken) error {
	filename := filepath.Join(ts.configDir, fmt.Sprintf("%s-token.json", token.Provider))

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token metadata: %w", err)
	}

	
	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write token metadata: %w", err)
	}

	return nil
}


func (ts *TokenStorage) loadMetadata(providerName string) (*StoredToken, error) {
	filename := filepath.Join(ts.configDir, fmt.Sprintf("%s-token.json", providerName))

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no token metadata found for %s", providerName)
		}
		return nil, fmt.Errorf("failed to read token metadata: %w", err)
	}

	var token StoredToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token metadata: %w", err)
	}

	return &token, nil
}


func StoreAPIKey(providerName, apiKey string) error {
	key := fmt.Sprintf("%s-api-key", providerName)
	return keyring.Set(serviceName, key, apiKey)
}


func LoadAPIKey(providerName string) (string, error) {
	key := fmt.Sprintf("%s-api-key", providerName)
	apiKey, err := keyring.Get(serviceName, key)
	if err == keyring.ErrNotFound {
		return "", fmt.Errorf("no API key found for %s", providerName)
	}
	return apiKey, err
}


func DeleteAPIKey(providerName string) error {
	key := fmt.Sprintf("%s-api-key", providerName)
	return keyring.Delete(serviceName, key)
}
