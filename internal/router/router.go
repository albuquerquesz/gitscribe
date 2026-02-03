package router

import (
	"context"
	"fmt"

	"github.com/albuquerquesz/gitscribe/internal/agents"
	"github.com/albuquerquesz/gitscribe/internal/config"
)

type Router struct {
	config  *config.Config
	factory *agents.Factory
}

func NewRouter(cfg *config.Config) *Router {
	return &Router{
		config:  cfg,
		factory: agents.NewFactory(),
	}
}

func (r *Router) RouteRequest(ctx context.Context, agentName string, messages []agents.Message, options agents.RequestOptions) (*agents.Response, error) {
	profile, err := r.config.GetAgentByName(agentName)
	if err != nil {
		return nil, fmt.Errorf("agent profile not found: %w", err)
	}

	client, err := r.factory.CreateClient(*profile)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for agent %s: %w", agentName, err)
	}

	return client.SendMessage(ctx, messages, options)
}