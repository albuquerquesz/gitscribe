package providers

import (
	"context"

	"github.com/albuquerquesz/gitscribe/internal/auth"
)

const (
	openAIAuthEndpoint  = "https://openai.com/oauth/authorize"
	openAITokenEndpoint = "https://api.openai.com/oauth/token"
)

var OpenAIScopes = []string{
	"user.read",
	"models.read",
	"completions.write",
}

type OpenAIProvider struct {
	baseURL string
}

func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{
		baseURL: "https://api.openai.com",
	}
}

func (o *OpenAIProvider) Name() string {
	return "openai"
}

func (o *OpenAIProvider) AuthorizationEndpoint() string {
	return openAIAuthEndpoint
}

func (o *OpenAIProvider) TokenEndpoint() string {
	return openAITokenEndpoint
}

func (o *OpenAIProvider) Scopes() []string {
	return OpenAIScopes
}

func (o *OpenAIProvider) ClientID() string {
	return "openai-public-client"
}

func (o *OpenAIProvider) SupportsPKCE() bool {
	return true
}

func (o *OpenAIProvider) APIKeyEndpoint() string {
	return o.baseURL + "/v1/api-keys"
}

func (o *OpenAIProvider) GenerateAPIKey(ctx context.Context, accessToken string) (string, error) {
	return accessToken, nil
}

var _ auth.Provider = (*OpenAIProvider)(nil)