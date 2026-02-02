package router

import (
	"context"
	"fmt"
	"strings"

	"github.com/albuquerquesz/gitscribe/internal/agents"
	"github.com/albuquerquesz/gitscribe/internal/config"
)

// Strategy defines how to select an agent
type Strategy string

const (
	// StrategyDefault always uses the default agent
	StrategyDefault Strategy = "default"

	// StrategyAuto automatically selects based on task complexity
	StrategyAuto Strategy = "auto"

	// StrategyRoundRobin cycles through available agents
	StrategyRoundRobin Strategy = "round-robin"

	// StrategyPriority uses priority-based selection
	StrategyPriority Strategy = "priority"

	// StrategyFallback tries default first, falls back on failure
	StrategyFallback Strategy = "fallback"
)

// RequestContext contains information about the request
type RequestContext struct {
	TaskType       string
	Complexity     string // low, medium, high
	TokenCount     int
	Requires       []string // reasoning, code, creative, etc.
	UserPrompt     string
	PreferredAgent string
}

// Router handles agent selection and request routing
type Router struct {
	config       *config.Config
	factory      *agents.Factory
	strategy     Strategy
	clients      map[string]agents.Client
	currentIndex int // For round-robin
}

// NewRouter creates a new request router
func NewRouter(cfg *config.Config, strategy Strategy) *Router {
	return &Router{
		config:   cfg,
		factory:  agents.NewFactory(),
		strategy: strategy,
		clients:  make(map[string]agents.Client),
	}
}

// SetStrategy changes the routing strategy
func (r *Router) SetStrategy(strategy Strategy) {
	r.strategy = strategy
}

// GetClient returns or creates a client for an agent
func (r *Router) GetClient(agentName string) (agents.Client, error) {
	// Return cached client if available
	if client, ok := r.clients[agentName]; ok {
		return client, nil
	}

	// Get agent profile
	profile, err := r.config.GetAgentByName(agentName)
	if err != nil {
		return nil, err
	}

	// Create new client
	client, err := r.factory.CreateClient(*profile)
	if err != nil {
		return nil, err
	}

	// Cache the client
	r.clients[agentName] = client
	return client, nil
}

// SelectAgent determines which agent to use based on strategy and context
func (r *Router) SelectAgent(ctx RequestContext) (*config.AgentProfile, error) {
	// If user specified a preferred agent, use it
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

// selectDefault returns the default agent
func (r *Router) selectDefault() (*config.AgentProfile, error) {
	return r.config.GetDefaultAgent()
}

// selectAuto chooses based on request context
func (r *Router) selectAuto(ctx RequestContext) (*config.AgentProfile, error) {
	enabled := r.config.ListEnabledAgents()
	if len(enabled) == 0 {
		return nil, fmt.Errorf("no enabled agents available")
	}

	// Apply routing rules from config
	for _, rule := range r.config.Routing {
		profile, err := r.config.GetAgentByName(rule.AgentProfile)
		if err != nil || !profile.Enabled {
			continue
		}

		if r.matchesRule(ctx, rule) {
			return profile, nil
		}
	}

	// Fallback: select based on complexity and capabilities
	return r.selectByComplexity(ctx, enabled)
}

// matchesRule checks if a request matches a routing rule
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

// compareInt compares an integer value
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

// contains checks if a string slice contains a value
func (r *Router) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// selectByComplexity chooses agent based on complexity
func (r *Router) selectByComplexity(ctx RequestContext, agents []config.AgentProfile) (*config.AgentProfile, error) {
	switch ctx.Complexity {
	case "high":
		// Prefer agents with larger context windows or better reasoning
		for _, agent := range agents {
			if agent.MaxTokens >= 4096 {
				return &agent, nil
			}
		}
	case "low":
		// Prefer fast, cheap agents
		for _, agent := range agents {
			if agent.Provider == config.ProviderGroq || agent.Provider == config.ProviderOllama {
				return &agent, nil
			}
		}
	}

	// Default: return first enabled
	if len(agents) > 0 {
		return &agents[0], nil
	}

	return nil, fmt.Errorf("no suitable agent found")
}

// selectRoundRobin cycles through agents
func (r *Router) selectRoundRobin() (*config.AgentProfile, error) {
	enabled := r.config.ListEnabledAgents()
	if len(enabled) == 0 {
		return nil, fmt.Errorf("no enabled agents available")
	}

	agent := &enabled[r.currentIndex%len(enabled)]
	r.currentIndex++
	return agent, nil
}

// selectPriority selects based on priority
func (r *Router) selectPriority() (*config.AgentProfile, error) {
	enabled := r.config.ListEnabledAgents()
	if len(enabled) == 0 {
		return nil, fmt.Errorf("no enabled agents available")
	}

	// Sort by priority (lower number = higher priority)
	var highest *config.AgentProfile
	for i := range enabled {
		if highest == nil || enabled[i].Priority < highest.Priority {
			highest = &enabled[i]
		}
	}

	return highest, nil
}

// selectFallback tries default first, falls back to others on failure
func (r *Router) selectFallback(ctx RequestContext) (*config.AgentProfile, error) {
	// First try the default or auto-selected agent
	defaultAgent, err := r.selectAuto(ctx)
	if err != nil {
		return nil, err
	}

	// Check if it's available
	client, err := r.GetClient(defaultAgent.Name)
	if err == nil && client.IsAvailable() {
		return defaultAgent, nil
	}

	// Fall back to next available agent
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

// RouteRequest routes a request to the appropriate agent
func (r *Router) RouteRequest(ctx context.Context, reqCtx RequestContext, messages []agents.Message, options agents.RequestOptions) (*agents.Response, error) {
	// Select agent
	profile, err := r.SelectAgent(reqCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to select agent: %w", err)
	}

	// Get client
	client, err := r.GetClient(profile.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for agent %s: %w", profile.Name, err)
	}

	// Send request
	resp, err := client.SendMessage(ctx, messages, options)
	if err != nil {
		return nil, fmt.Errorf("agent %s failed: %w", profile.Name, err)
	}

	return resp, nil
}

// Close cleans up all clients
func (r *Router) Close() error {
	for _, client := range r.clients {
		if err := client.Close(); err != nil {
			// Log error but continue closing others
			fmt.Printf("Error closing client: %v\n", err)
		}
	}
	r.clients = make(map[string]agents.Client)
	return nil
}

// ListAvailableAgents returns all available (enabled) agents
func (r *Router) ListAvailableAgents() []config.AgentProfile {
	return r.config.ListEnabledAgents()
}

// GetAgentInfo returns information about a specific agent
func (r *Router) GetAgentInfo(name string) (*config.AgentProfile, bool, error) {
	profile, err := r.config.GetAgentByName(name)
	if err != nil {
		return nil, false, err
	}

	// Check if client exists and is available
	if client, ok := r.clients[name]; ok {
		return profile, client.IsAvailable(), nil
	}

	return profile, false, nil
}
