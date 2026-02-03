package providers

import (
	"github.com/albuquerquesz/gitscribe/internal/auth"
)

type OpenAIProvider struct{}

func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{}
}

func (o *OpenAIProvider) Name() string {
	return "openai"
}

var _ auth.Provider = (*OpenAIProvider)(nil)
