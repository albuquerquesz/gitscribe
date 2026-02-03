package style

import (
	"fmt"
	"strings"
	"time"

	"github.com/albuquerquesz/gitscribe/internal/catalog"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	InfoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	WarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	TitleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Bold(true).MarginBottom(1)

	ProviderHeaderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Bold(true).PaddingLeft(1)
	ModelItemStyle      = lipgloss.NewStyle().PaddingLeft(4)
	DimStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)
)

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
|  $$$$$$/
 \______/
`
	fmt.Println(SuccessStyle.Render(ascii))
	time.Sleep(500 * time.Millisecond)
}

func GroupedSelect(title string, groups map[string][]huh.Option[string]) (string, error) {
	var options []huh.Option[string]

	providers := []string{"openai", "anthropic", "groq", "opencode", "ollama"}

	for _, p := range providers {
		models, ok := groups[p]
		if !ok || len(models) == 0 {
			continue
		}

		options = append(options, huh.NewOption(ProviderHeaderStyle.Render("── "+strings.ToUpper(p)+" ──"), "header:"+p))

		for _, opt := range models {
			label := ModelItemStyle.Render(opt.Key)
			options = append(options, huh.NewOption(label, opt.Value))
		}
	}

	var selected string
	for {
		err := huh.NewSelect[string]().
			Title(title).
			Options(options...).
			Value(&selected).
			Run()
		if err != nil {
			return "", err
		}

		if strings.HasPrefix(selected, "header:") {
			continue
		}

		return selected, nil
	}
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

