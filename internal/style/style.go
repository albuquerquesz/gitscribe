package style

import (
	"fmt"
	"strings"
	"time"

	"github.com/albuquerquesz/gitscribe/internal/catalog"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

var (
	Purple = lipgloss.Color("#7D56F4")
	Green  = lipgloss.Color("#04B575")
	Red    = lipgloss.Color("#EF4444")
	Amber  = lipgloss.Color("#F59E0B")
	Blue   = lipgloss.Color("#3B82F6")
	Grey   = lipgloss.Color("#6B7280")

	SuccessStyle = lipgloss.NewStyle().Foreground(Green).Bold(true)
	ErrorStyle   = lipgloss.NewStyle().Foreground(Red).Bold(true)
	InfoStyle    = lipgloss.NewStyle().Foreground(Blue)
	WarningStyle = lipgloss.NewStyle().Foreground(Amber)
	TitleStyle   = lipgloss.NewStyle().Foreground(Purple).Bold(true).MarginBottom(1)

	ProviderHeaderStyle = lipgloss.NewStyle().Foreground(Purple).Bold(true).PaddingLeft(1)
	ModelItemStyle      = lipgloss.NewStyle().PaddingLeft(4)
	DimStyle            = lipgloss.NewStyle().Foreground(Grey)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Purple).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)
)

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
	fmt.Printf("⏳ %s...\n", msg)
	return &SimpleSpinner{message: msg}
}

func ConfirmAction(msg string) bool {
	var confirm bool
	err := huh.NewConfirm().
		Title(msg).
		Affirmative("Yes").
		Negative("No").
		Value(&confirm).
		Run()

	return err == nil && confirm
}

func GetASCIIName() {
	ascii := `
           /$$   /$$                                  /$$ /$$ 
          |__/  | $$
  /$$$$$$  /$$ /$$$$$$   /$$$$$$$  /$$$$$$$  /$$$$$$  /$$| $$$$$$$   /$$$$$$ 
 /$$__  $$| $$|_  $$_/  /$$_____/ /$$_____/ /$$__  $$| $$| $$__  $$ /$$__  $$
| $$  \ $$| $$  | $$   |  $$$$$$ | $$      | $$  \__/| $$| $$  \ $$| $$$$$$$
| $$  | $$| $$  | $$ /$\____  $$| $$      | $$      | $$| $$  | $$| $$_____/
|  $$$$$$$| $$  |  $$$$//$$$$$$$/|  $$$$$$$| $$      | $$| $$$$$$$/|  $$$$$$$
 \____  $$|__/   \___/ |_______/  \_______/|__/      |__/|_______/  \_______/
 /$$  \ $$
|  $$$$$$
 \______/
`
	lines := strings.Split(ascii, "\n")
	startColor, _ := colorful.Hex("#00BFFF")
	endColor, _ := colorful.Hex("#7D56F4")

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			fmt.Println()
			continue
		}
		t := float64(i) / float64(len(lines)-1)
		c := startColor.BlendLuv(endColor, t).Hex()
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Render(line))
	}
	time.Sleep(500 * time.Millisecond)
}

func formatProviderName(p string) string {
	pLower := strings.ToLower(p)
	if pLower == "groq" {
		return "GROQ"
	}
	if pLower == "openai" {
		return "OpenAI"
	}
	if pLower == "opencode" {
		return "OpenCode Zen"
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

func SelectModel(manager *catalog.CatalogManager) (*catalog.Model, error) {
	var selectedProvider string
	var selectedModelID string

	providers := manager.ListProviders()
	var providerOpts []huh.Option[string]
	for _, p := range providers {
		providerOpts = append(providerOpts, huh.NewOption(formatProviderName(p), p))
	}

	if len(providerOpts) > 0 {
		selectedProvider = providerOpts[0].Value
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Provider").
				Options(providerOpts...).
				Value(&selectedProvider),

			huh.NewSelect[string]().
				Title("Select Model").
				OptionsFunc(func() []huh.Option[string] {
					return getModelOptions(manager, selectedProvider)
				}, &selectedProvider).
				Value(&selectedModelID),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	return manager.GetModel(selectedModelID)
}

func Prompt(label string) (string, error) {
	var input string
	err := huh.NewInput().
		Title(label).
		EchoMode(huh.EchoModePassword).
		Value(&input).
		Run()
	return input, err
}

func StringMask(str string) string {
	if len(str) <= 8 {
		return "********"
	}
	return str[:8] + "********"
}

func Info(msg string) {
	fmt.Println(InfoStyle.Render("ℹ " + msg))
}

func Success(msg string) {
	fmt.Println(SuccessStyle.Render("✓ " + msg))
}

func Error(msg string) {
	fmt.Println(ErrorStyle.Render("✖ " + msg))
}

func Warning(msg string) {
	fmt.Println(WarningStyle.Render("⚠ " + msg))
}

func Box(title, content string) {
	styledContent := fmt.Sprintf("%s\n\n%s", TitleStyle.Render(title), content)
	fmt.Println(BoxStyle.Render(styledContent))
}

func InteractiveConfirm(msg string) bool {
	return ConfirmAction(msg)
}