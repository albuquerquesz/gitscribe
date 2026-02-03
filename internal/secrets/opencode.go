package secrets

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type OpenCodeAuthEntry struct {
	Type    string `json:"type"`
	Key     string `json:"key,omitempty"`
	Access  string `json:"access,omitempty"`
	Refresh string `json:"refresh,omitempty"`
	Expires int64  `json:"expires,omitempty"`
}

type OpenCodeAuth map[string]OpenCodeAuthEntry

func GetOpenCodeAuthPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "opencode", "auth.json"), nil
}

func LoadOpenCodeAuth() (OpenCodeAuth, error) {
	path, err := GetOpenCodeAuthPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read opencode auth: %w", err)
	}

	var auth OpenCodeAuth
	if err := json.Unmarshal(data, &auth); err != nil {
		return nil, fmt.Errorf("failed to parse opencode auth: %w", err)
	}

	return auth, nil
}

func (o OpenCodeAuth) GetAPIKey(provider string) (string, bool) {
	if o == nil {
		return "", false
	}

	mappedProvider := o.MapProvider(provider)
	entry, ok := o[mappedProvider]
	if !ok {
		return "", false
	}

	if entry.Type == "api" && entry.Key != "" {
		return entry.Key, true
	}

	if entry.Type == "oauth" && entry.Access != "" {
		if entry.Expires > 0 && time.Now().UnixMilli() > entry.Expires {
			return "", false
		}
		return entry.Access, true
	}

	return "", false
}

func (o OpenCodeAuth) IsTokenExpired(provider string) bool {
	if o == nil {
		return true
	}

	mappedProvider := o.MapProvider(provider)
	entry, ok := o[mappedProvider]
	if !ok {
		return true
	}

	if entry.Type != "oauth" {
		return false
	}

	if entry.Expires <= 0 {
		return false
	}

	return time.Now().UnixMilli() > entry.Expires
}

func (o OpenCodeAuth) GetTokenExpiry(provider string) (time.Time, bool) {
	if o == nil {
		return time.Time{}, false
	}

	mappedProvider := o.MapProvider(provider)
	entry, ok := o[mappedProvider]
	if !ok {
		return time.Time{}, false
	}

	if entry.Type != "oauth" || entry.Expires <= 0 {
		return time.Time{}, false
	}

	return time.UnixMilli(entry.Expires), true
}

func (o OpenCodeAuth) IsTokenExpiringSoon(provider string, threshold time.Duration) bool {
	expiry, ok := o.GetTokenExpiry(provider)
	if !ok {
		return false
	}

	return time.Until(expiry) < threshold
}

func (o OpenCodeAuth) MapProvider(provider string) string {
	mapping := map[string]string{
		"anthropic":  "anthropic",
		"openai":     "openai",
		"groq":       "groq",
		"opencode":   "opencode",
		"gemini":     "google",
		"google":     "google",
		"openrouter": "openrouter",
		"ollama":     "ollama",
	}
	if mapped, ok := mapping[provider]; ok {
		return mapped
	}
	return provider
}

func (o OpenCodeAuth) ListProviders() []string {
	if o == nil {
		return nil
	}
	providers := make([]string, 0, len(o))
	for p := range o {
		providers = append(providers, p)
	}
	return providers
}
