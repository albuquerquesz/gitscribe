package ai

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

func NewClient(apiKey string) *openai.Client {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.groq.com/openai/v1"

	return openai.NewClientWithConfig(config)
}

func ValidateGroqToken(apiKey string) (bool, error) {
	if apiKey == "" {
		return false, nil
	}
	client := NewClient(apiKey)
	_, err := client.ListModels(context.Background())
	if err != nil {
		return false, err
	}

	return true, nil
}
