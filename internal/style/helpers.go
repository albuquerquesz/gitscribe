package style

import (
	"context"
	"strings"

	"github.com/albuquerquesz/gitscribe/internal/catalog"
	"github.com/charmbracelet/huh"
	spinner "github.com/charmbracelet/huh/spinner"
)

func formatProviderName(p string) string {
	pLower := strings.ToLower(p)
	if pLower == "groq" {
		return "GROQ"
	}
	if pLower == "openai" {
		return "OpenAI"
	}
	if pLower == "opencode" {
		return "OpenCode"
	}
	if len(p) > 0 {
		return strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	return p
}

func getModelOptions(manager *catalog.CatalogManager, provider string) []huh.Option[string] {
	models := manager.GetModelsByProvider(provider)
	var opts []huh.Option[string]
	for _, mod := range models {
		opts = append(opts, huh.NewOption(mod.Name, mod.ID))
	}
	if len(opts) == 0 {
		opts = append(opts, huh.NewOption("No models available", ""))
	}
	return opts
}

func Spinner(ctx context.Context, title string) *spinner.Spinner {
	return spinner.New().Title(title).Context(ctx)
}

func RunWithSpinner(title string, action func() error) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := Spinner(ctx, title)
	done := make(chan error, 1)
	go func() {
		done <- s.Run()
	}()

	err := action()
	cancel()
	<-done

	return err
}
