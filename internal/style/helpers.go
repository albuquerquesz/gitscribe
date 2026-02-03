package style

import (
	"fmt"
	"strings"

	"github.com/albuquerquesz/gitscribe/internal/catalog"
	"github.com/charmbracelet/huh"
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

type SimpleSpinner struct {
	message string
}

func (s *SimpleSpinner) Stop() {
}

func (s *SimpleSpinner) Success(msg string) {
	fmt.Println(SuccessStyle.Render("✓ " + msg))
}

func (s *SimpleSpinner) Fail(msg string) {
	fmt.Println(ErrorStyle.Render("✖ " + msg))
}

func (s *SimpleSpinner) Warning(msg string) {
	fmt.Println(WarningStyle.Render("⚠ " + msg))
}

func (s *SimpleSpinner) UpdateText(msg string) {
	s.message = msg
}

func Spinner(msg string) *SimpleSpinner {
	fmt.Printf("%s...\n", msg)
	return &SimpleSpinner{message: msg}
}
