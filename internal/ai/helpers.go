package ai

import "context"

type responseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type choice struct {
	Index        int             `json:"index"`
	Message      responseMessage `json:"message"`
	FinishReason string          `json:"finish_reason"`
}

type APIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []choice `json:"choices"`
}

func ValidateToken(apiKey string) bool {
	if apiKey == "" {
		return false
	}

	client := newClient(apiKey)
	_, err := client.ListModels(context.Background())
	if err != nil {
		return false
	}

	return true
}
