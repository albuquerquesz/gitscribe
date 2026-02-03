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
		ID:          "kimi-2.5",
		Provider:    "opencode",
		Name:        "Kimi 2.5",
		Description: "Long context specialist from OpenCode Zen",
	},
	{
		ID:          "mini-pickle",
		Provider:    "opencode",
		Name:        "Mini Pickle",
		Description: "Fast and lightweight coding assistant",
	},
	{
		ID:          "glm",
		Provider:    "opencode",
		Name:        "GLM",
		Description: "Powerful General Language Model",
	},
	{
		ID:       "llama-3.3-70b-versatile",
		Provider: "groq",
		Name:     "Llama 3.3 70B Versatile",
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
		BaseURL:    "https://api.opencode.com/v1",
		AuthMethod: AuthMethodAPIKey,
	},
}

func GetProviderConfig(name string) (ProviderConfig, bool) {
	config, ok := ProviderConfigs[name]
	return config, ok
}
