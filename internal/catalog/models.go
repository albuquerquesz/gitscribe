package catalog

import (
	"fmt"
)

type Model struct {
	ID          string `json:"id" yaml:"id"`
	Provider    string `json:"provider" yaml:"provider"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type ProviderConfig struct {
	Name       string     `json:"name" yaml:"name"`
	BaseURL    string     `json:"base_url" yaml:"base_url"`
	AuthMethod AuthMethod `json:"auth_method" yaml:"auth_method"`
}

type AuthMethod string

const (
	AuthMethodAPIKey AuthMethod = "api_key"
	AuthMethodBearer AuthMethod = "bearer"
	AuthMethodNone   AuthMethod = "none"
)

type ModelCatalog struct {
	Models    []Model          `json:"models" yaml:"models"`
	Providers []ProviderConfig `json:"providers" yaml:"providers"`
}

func (c *ModelCatalog) GetModelByID(id string) (*Model, error) {
	for i := range c.Models {
		if c.Models[i].ID == id {
			return &c.Models[i], nil
		}
	}
	return nil, fmt.Errorf("model not found: %s", id)
}
