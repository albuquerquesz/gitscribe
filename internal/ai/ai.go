package ai

import (
	"context"
	"fmt"

	"github.com/albqvictor1508/gitscribe/internal/store"
	openai "github.com/sashabaranov/go-openai"
)

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

func SendPrompt(diff string) (string, error) {
	ctx := fmt.Sprintf(
		"Analyze the following git diff and generate a commit message. "+
			"The message must follow the Conventional Commits standard. "+
			"Your response should contain *only* the commit message, without any additional text, explanations, or markdown formatting. "+
			"Focus on the primary purpose of the changes and be concise. "+
			"Do not include file names, line numbers, or the diff itself in the output. "+
			"Here is the diff:\n%v",
		diff,
	)

	apiKey, err := store.Get()
	if err != nil {
		return "", fmt.Errorf("error to get api key: %w", err)
	}

	return requestAI(apiKey, ctx)
}

func requestAI(apiKey, ctx string) (string, error) {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.groq.com/openai/v1"

	client := openai.NewClientWithConfig(config)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "llama-3.3-70b-versatile",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    "system",
					Content: "You are a commit message generator, and you have to generate commit messages in the Conventional Commits pattern",
				},
				{
					Role:    "user",
					Content: ctx,
				},
			},
		},
	)
	if err != nil {
		fmt.Println("Erro:", err)
		return "", fmt.Errorf("error: %v", err)
	}

	msg := resp.Choices[0].Message.Content
	return msg, nil
}
