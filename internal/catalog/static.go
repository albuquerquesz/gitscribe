package catalog

var StaticModels = []Model{
	{
		ID:       "claude-3-5-sonnet-20241022",
		Provider: "anthropic",
		Name:     "Claude 3.5 Sonnet",
	},
	{
		ID:       "claude-3-5-haiku-20241022",
		Provider: "anthropic",
		Name:     "Claude 3.5 Haiku",
	},
	{
		ID:       "gpt-4o",
		Provider: "openai",
		Name:     "GPT-4o",
	},
	{
		ID:       "gpt-4o-mini",
		Provider: "openai",
		Name:     "GPT-4o Mini",
	},
	{
		ID:          "kimi-k2.5-free",
		Provider:    "opencode",
		Name:        "Kimi 2.5 Free",
		Description: "Long context specialist from OpenCode Zen",
	},
	{
		ID:          "minimax-m2.1-free",
		Provider:    "opencode",
		Name:        "MiniMax M2.1 Free",
		Description: "Fast and lightweight coding assistant",
	},
	{
		ID:          "glm-4.7-free",
		Provider:    "opencode",
		Name:        "GLM 4.7 Free",
		Description: "Powerful General Language Model",
	},
	{
		ID:       "llama-3.3-70b-versatile",
		Provider: "groq",
		Name:     "Llama 3.3 70B Versatile",
	},
	{
		ID:       "openai/gpt-oss-120b",
		Provider: "groq",
		Name:     "OpenAI GPT OSS 120b",
	},
	{
		ID:          "moonshotai/kimi-k2.5",
		Provider:    "hackclub",
		Name:        "Kimi K2.5",
		Description: "Reasoning model via Hack Club",
	},
	{
		ID:          "qwen/qwen-2.5-72b-instruct",
		Provider:    "hackclub",
		Name:        "Qwen 2.5 72B",
		Description: "Powerful open model via Hack Club",
	},
}

func GetStaticModels() []Model {
	return StaticModels
}

var ProviderConfigs = map[string]ProviderConfig{
	"anthropic": {
		Name:       "anthropic",
		BaseURL:    "https://api.anthropic.com/v1",
		AuthMethod: AuthMethodAPIKey,
	},
	"openai": {
		Name:       "openai",
		BaseURL:    "https://api.openai.com/v1",
		AuthMethod: AuthMethodBearer,
	},
	"groq": {
		Name:       "groq",
		BaseURL:    "https://api.groq.com/openai/v1",
		AuthMethod: AuthMethodBearer,
	},
	"opencode": {
		Name:       "opencode",
		BaseURL:    "https://opencode.ai/zen/v1",
		AuthMethod: AuthMethodAPIKey,
	},
	"hackclub": {
		Name:       "hackclub",
		BaseURL:    "https://ai.hackclub.com/proxy/v1",
		AuthMethod: AuthMethodBearer,
	},
}

func GetProviderConfig(name string) (ProviderConfig, bool) {
	config, ok := ProviderConfigs[name]
	return config, ok
}
