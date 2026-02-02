package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/albuquerquesz/gitscribe/internal/agents"
	"github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/router"
	"github.com/albuquerquesz/gitscribe/internal/secrets"
	"github.com/spf13/cobra"
)

var multiAgentCmd = &cobra.Command{
	Use:   "ask",
	Short: "Send a prompt to an AI agent",
	Long:  "Sends a prompt to an AI agent using the multi-agent routing system",
	Example: `  # Use default agent
  gs ask "What is the weather today?"

  # Use specific agent
  gs ask "Explain quantum computing" --agent claude-sonnet

  # Use auto-routing based on complexity
  gs ask "Write a complex algorithm" --strategy auto`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("please provide a prompt")
		}
		return runMultiAgent(args[0])
	},
}

var (
	agentName      string
	routerStrategy string
	agentList      bool
)

func init() {
	multiAgentCmd.Flags().StringVarP(&agentName, "agent", "a", "", "Agent profile to use (overrides strategy)")
	multiAgentCmd.Flags().StringVarP(&routerStrategy, "strategy", "s", "default", "Routing strategy: default, auto, round-robin, priority, fallback")
	multiAgentCmd.Flags().BoolVar(&agentList, "list", false, "List available agents")

	rootCmd.AddCommand(multiAgentCmd)
}

func runMultiAgent(prompt string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if agentList {
		return listAgents(cfg)
	}

	strategy := router.Strategy(routerStrategy)
	r := router.NewRouter(cfg, strategy)
	defer r.Close()

	reqCtx := router.RequestContext{
		UserPrompt:     prompt,
		PreferredAgent: agentName,
		Complexity:     detectComplexity(prompt),
	}

	messages := []agents.Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	options := agents.RequestOptions{
		Temperature: 0.7,
		Timeout:     60 * time.Second,
	}

	fmt.Println("🤖 Sending request...")

	ctx := context.Background()
	resp, err := r.RouteRequest(ctx, reqCtx, messages, options)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	fmt.Printf("\n📤 Response (Model: %s):\n", resp.Model)
	fmt.Println(resp.Content)
	fmt.Printf("\n📊 Tokens used: %d (prompt: %d, completion: %d)\n",
		resp.Usage.TotalTokens, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)

	return nil
}

func listAgents(cfg *config.Config) error {
	fmt.Println("📋 Configured Agents:")
	fmt.Println()

	for _, agent := range cfg.Agents {
		status := "❌ Disabled"
		if agent.Enabled {
			status = "✅ Enabled"
		}

		defaultMarker := ""
		if cfg.Global.DefaultAgent == agent.Name {
			defaultMarker = " (default)"
		}

		fmt.Printf("  %s%s\n", agent.Name, defaultMarker)
		fmt.Printf("    Provider: %s\n", agent.Provider)
		fmt.Printf("    Model: %s\n", agent.Model)
		fmt.Printf("    Status: %s\n", status)
		fmt.Printf("    Priority: %d\n", agent.Priority)

		keyMgr := secrets.NewAgentKeyManager()
		hasKey := keyMgr.KeyExists(keyMgr.GetAgentKeyName(agent.Name))
		keyStatus := "❌ Not configured"
		if hasKey {
			keyStatus = "✅ Configured"
		}
		fmt.Printf("    API Key: %s\n", keyStatus)
		fmt.Println()
	}

	return nil
}

func detectComplexity(prompt string) string {
	length := len(prompt)

	if length < 100 {
		return "low"
	} else if length < 500 {
		return "medium"
	}
	return "high"
}
