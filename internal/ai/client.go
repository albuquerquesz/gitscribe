package ai

import (
	openai "github.com/sashabaranov/go-openai"
)

func newClient(apiKey string) *openai.Client {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.groq.com/openai/v1"

	return openai.NewClientWithConfig(config)
}
