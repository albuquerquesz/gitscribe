package providers

import (
	"github.com/albuquerquesz/gitscribe/internal/auth"
)

type AnthropicProvider struct{}

func NewAnthropicProvider() *AnthropicProvider {
	return &AnthropicProvider{}
}

func (a *AnthropicProvider) Name() string {
	return "anthropic"
}

var _ auth.Provider = (*AnthropicProvider)(nil)
