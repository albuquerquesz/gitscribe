package ai

import (
	"context"
	"fmt"

	"github.com/albuquerquesz/gitscribe/internal/store"
	openai "github.com/sashabaranov/go-openai"
)

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
	client := newClient(apiKey)

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
		return "", fmt.Errorf("error: %v", err)
	}

	msg := resp.Choices[0].Message.Content
	return msg, nil
}
