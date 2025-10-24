package internal

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

type ResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Choice struct {
	Index        int             `json:"index"`
	Message      ResponseMessage `json:"message"`
	FinishReason string          `json:"finish_reason"`
}

type APIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

func SendPrompt(ctx string) (string, error) {
	godotenv.Load()

	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return "", errors.New("Missing GROQ_API_KEY env...")
	}

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
		return "", fmt.Errorf("Erro: %v", err)
	}

	msg := resp.Choices[0].Message.Content
	return msg, nil
}
