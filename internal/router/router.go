package router

import (
	"context"
	"fmt"
	"strings"

	"github.com/albuquerquesz/gitscribe/internal/agents"
	"github.com/albuquerquesz/gitscribe/internal/config"
)


type Strategy string

const (
	
	StrategyDefault Strategy = "default"

	
	StrategyAuto Strategy = "auto"

	
	StrategyRoundRobin Strategy = "round-robin"

	
	StrategyPriority Strategy = "priority"

	
	StrategyFallback Strategy = "fallback"
)


type RequestContext struct {
	TaskType       string
	Complexity     string 
	TokenCount     int
	Requires       []string 
	UserPrompt     string
	PreferredAgent string
}


type Router struct {
	config       *config.Config
	factory      *agents.Factory
	strategy     Strategy
	clients      map[string]agents.Client
	currentIndex int 
}


func NewRouter(cfg *config.Config, strategy Strategy) *Router {
	return &Router{
		config:   cfg,
		factory:  agents.NewFactory(),
		strategy: strategy,
		clients:  make(map[string]agents.Client),
	}
}


func (r *Router) SetStrategy(strategy Strategy) {
	r.strategy = strategy
}


func (r *Router) GetClient(agentName string) (agents.Client, error) {
	
	if client, ok := r.clients[agentName]; ok {
		return client, nil
	}

	
	profile, err := r.config.GetAgentByName(agentName)
	if err != nil {
		return nil, err
	}

	
	client, err := r.factory.CreateClient(*profile)
	if err != nil {
		return nil, err
	}

	
	r.clients[agentName] = client
	return client, nil
}


func (r *Router) SelectAgent(ctx RequestContext) (*config.AgentProfile, error) {
	
	if ctx.PreferredAgent != "" {
		profile, err := r.config.GetAgentByName(ctx.PreferredAgent)
		if err != nil {
			return nil, fmt.Errorf("preferred agent not found: %w", err)
		}
		if !profile.Enabled {
			return nil, fmt.Errorf("preferred agent is not enabled: %s", ctx.PreferredAgent)
		}
		return profile, nil
	}

	switch r.strategy {
	case StrategyDefault:
		return r.selectDefault()

	case StrategyAuto:
		return r.selectAuto(ctx)

	case StrategyRoundRobin:
		return r.selectRoundRobin()

	case StrategyPriority:
		return r.selectPriority()

	case StrategyFallback:
		return r.selectFallback(ctx)

	default:
		return r.selectDefault()
	}
}


func (r *Router) selectDefault() (*config.AgentProfile, error) {
	return r.config.GetDefaultAgent()
}


func (r *Router) selectAuto(ctx RequestContext) (*config.AgentProfile, error) {
	enabled := r.config.ListEnabledAgents()
	if len(enabled) == 0 {
		return nil, fmt.Errorf("no enabled agents available")
	}

	
	for _, rule := range r.config.Routing {
		profile, err := r.config.GetAgentByName(rule.AgentProfile)
		if err != nil || !profile.Enabled {
			continue
		}

		if r.matchesRule(ctx, rule) {
			return profile, nil
		}
	}

	
	return r.selectByComplexity(ctx, enabled)
}


func (r *Router) matchesRule(ctx RequestContext, rule config.RoutingRule) bool {
	for _, condition := range rule.Conditions {
		parts := strings.Split(condition, " ")
		if len(parts) != 3 {
			continue
		}

		field := parts[0]
		operator := parts[1]
		value := parts[2]

		switch field {
		case "token_count":
			if !r.compareInt(ctx.TokenCount, operator, value) {
				return false
			}
		case "complexity":
			if ctx.Complexity != value {
				return false
			}
		case "reasoning":
			if !r.contains(ctx.Requires, "reasoning") && value == "required" {
				return false
			}
		}
	}

	return true
}


func (r *Router) compareInt(actual int, operator, value string) bool {
	var threshold int
	fmt.Sscanf(value, "%d", &threshold)

	switch operator {
	case "<":
		return actual < threshold
	case ">":
		return actual > threshold
	case "<=":
		return actual <= threshold
	case ">=":
		return actual >= threshold
	case "=":
		return actual == threshold
	default:
		return false
	}
}


func (r *Router) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}


func (r *Router) selectByComplexity(ctx RequestContext, agents []config.AgentProfile) (*config.AgentProfile, error) {
	switch ctx.Complexity {
	case "high":
		
		for _, agent := range agents {
			if agent.MaxTokens >= 4096 {
				return &agent, nil
			}
		}
	case "low":
		
		for _, agent := range agents {
			if agent.Provider == config.ProviderGroq || agent.Provider == config.ProviderOllama {
				return &agent, nil
			}
		}
	}

	
	if len(agents) > 0 {
		return &agents[0], nil
	}

	return nil, fmt.Errorf("no suitable agent found")
}


func (r *Router) selectRoundRobin() (*config.AgentProfile, error) {
	enabled := r.config.ListEnabledAgents()
	if len(enabled) == 0 {
		return nil, fmt.Errorf("no enabled agents available")
	}

	agent := &enabled[r.currentIndex%len(enabled)]
	r.currentIndex++
	return agent, nil
}


func (r *Router) selectPriority() (*config.AgentProfile, error) {
	enabled := r.config.ListEnabledAgents()
	if len(enabled) == 0 {
		return nil, fmt.Errorf("no enabled agents available")
	}

	
	var highest *config.AgentProfile
	for i := range enabled {
		if highest == nil || enabled[i].Priority < highest.Priority {
			highest = &enabled[i]
		}
	}

	return highest, nil
}


func (r *Router) selectFallback(ctx RequestContext) (*config.AgentProfile, error) {
	
	defaultAgent, err := r.selectAuto(ctx)
	if err != nil {
		return nil, err
	}

	
	client, err := r.GetClient(defaultAgent.Name)
	if err == nil && client.IsAvailable() {
		return defaultAgent, nil
	}

	
	enabled := r.config.ListEnabledAgents()
	for _, agent := range enabled {
		if agent.Name == defaultAgent.Name {
			continue
		}

		client, err := r.GetClient(agent.Name)
		if err == nil && client.IsAvailable() {
			return &agent, nil
		}
	}

	return nil, fmt.Errorf("no available agents found")
}


func (r *Router) RouteRequest(ctx context.Context, reqCtx RequestContext, messages []agents.Message, options agents.RequestOptions) (*agents.Response, error) {
	
	profile, err := r.SelectAgent(reqCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to select agent: %w", err)
	}

	
	client, err := r.GetClient(profile.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for agent %s: %w", profile.Name, err)
	}

	
	resp, err := client.SendMessage(ctx, messages, options)
	if err != nil {
		return nil, fmt.Errorf("agent %s failed: %w", profile.Name, err)
	}

	return resp, nil
}


func (r *Router) Close() error {
	for _, client := range r.clients {
		if err := client.Close(); err != nil {
			
			fmt.Printf("Error closing client: %v\n", err)
		}
	}
	r.clients = make(map[string]agents.Client)
	return nil
}


func (r *Router) ListAvailableAgents() []config.AgentProfile {
	return r.config.ListEnabledAgents()
}


func (r *Router) GetAgentInfo(name string) (*config.AgentProfile, bool, error) {
	profile, err := r.config.GetAgentByName(name)
	if err != nil {
		return nil, false, err
	}

	
	if client, ok := r.clients[name]; ok {
		return profile, client.IsAvailable(), nil
	}

	return profile, false, nil
}
