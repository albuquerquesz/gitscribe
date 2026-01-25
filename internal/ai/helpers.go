package ai

import (
	"context"
)

func ValidateToken(apiKey string) bool {
	if apiKey == "" {
		return false
	}

	client := newClient(apiKey)
	_, err := client.ListModels(context.Background())
	return err == nil
}
