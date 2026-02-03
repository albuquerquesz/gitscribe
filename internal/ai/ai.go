package ai

import (
	"context"
	"fmt"

	"github.com/albuquerquesz/gitscribe/internal/agents"
	"github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/router"
)

func SendPrompt(diff string, agentOverride string) (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	agent, err := cfg.GetDefaultAgent()
	if agentOverride != "" {
		agent, err = cfg.GetAgentByName(agentOverride)
	}

	if err != nil {
		return "", fmt.Errorf("no suitable agent found: %w", err)
	}

	r := router.NewRouter(cfg)

	prompt := fmt.Sprintf(
		"Analyze the following git diff and generate a commit message. "+
			"The message must follow the Conventional Commits standard. "+
			"Your response should contain *only* the commit message, without any additional text, explanations, or markdown formatting. "+
			"Focus on the primary purpose of the changes and be concise. "+
			"Do not include file names, line numbers, or the diff itself in the output. "+
			"Here is the diff:\n%v",
		diff,
	)

	messages := []agents.Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	options := agents.RequestOptions{
		Temperature: 0.7,
	}

	resp, err := r.RouteRequest(context.Background(), agent.Name, messages, options)
	if err != nil {
		return "", fmt.Errorf("ai request failed: %w", err)
	}

	return resp.Content, nil
}
