package auth

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "gitscribe-api-keys"
)

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

func IsAuthenticated(providerName string) (bool, error) {
	apiKey, err := LoadAPIKey(providerName)
	if err != nil || apiKey == "" {
		return false, nil
	}
	return true, nil
}
