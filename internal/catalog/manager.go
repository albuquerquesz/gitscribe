package catalog

import (
	"fmt"
)

type CatalogManager struct {
	apiKeyResolver func(provider string) (string, error)
}

func NewCatalogManager(resolver func(string) (string, error)) *CatalogManager {
	return &CatalogManager{apiKeyResolver: resolver}
}

func (cm *CatalogManager) ListProviders() []string {
	var list []string
	for k := range ProviderConfigs {
		list = append(list, k)
	}
	return list
}

func (cm *CatalogManager) GetModelsByProvider(provider string) ([]Model) {
	var models []Model
	for _, m := range StaticModels {
		if m.Provider == provider {
			models = append(models, m)
		}
	}
	return models
}

func (cm *CatalogManager) GetModel(id string) (*Model, error) {
	for i := range StaticModels {
		if StaticModels[i].ID == id {
			return &StaticModels[i], nil
		}
	}
	return nil, fmt.Errorf("model not found: %s", id)
}

func (cm *CatalogManager) GetProviderConfig(name string) (ProviderConfig, bool) {
	return GetProviderConfig(name)
}
