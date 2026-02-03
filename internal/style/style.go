package style

import (
	"fmt"
	"strings"

	"github.com/albuquerquesz/gitscribe/internal/catalog"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	White     = lipgloss.Color("#FFFFFF")
	LightGrey = lipgloss.Color("#E8E8E8")
	Grey      = lipgloss.Color("#A0A0A0")
	DarkGrey  = lipgloss.Color("#505050")
	Black     = lipgloss.Color("#000000")

	SuccessStyle = lipgloss.NewStyle().Foreground(Grey)
	ErrorStyle   = lipgloss.NewStyle().Foreground(DarkGrey)
	InfoStyle    = lipgloss.NewStyle().Foreground(LightGrey)
	WarningStyle = lipgloss.NewStyle().Foreground(Grey)
	TitleStyle   = lipgloss.NewStyle().Foreground(White).Bold(true).MarginBottom(1)

	ProviderHeaderStyle = lipgloss.NewStyle().Foreground(Grey).Bold(true).PaddingLeft(1)
	ModelItemStyle      = lipgloss.NewStyle().PaddingLeft(4)
	DimStyle            = lipgloss.NewStyle().Foreground(Grey)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(DarkGrey).
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
          |__/  | $$                                 |__/| $$                
  /$$$$$$  /$$ /$$$$$$   /$$$$$$$  /$$$$$$$  /$$$$$$  /$$| $$$$$$$   /$$$$$$ 
 /$$__  $$| $$|_  $$_/  /$$_____/ /$$_____/ /$$__  $$| $$| $$__  $$ /$$__  $$
| $$  \ $$| $$  | $$   |  $$$$$$ | $$      | $$  \__/| $$| $$  \ $$| $$$$$$$$
| $$  | $$| $$  | $$ /$$\____  $$| $$      | $$      | $$| $$  | $$| $$_____/
|  $$$$$$$| $$  |  $$$$//$$$$$$$/|  $$$$$$$| $$      | $$| $$$$$$$/|  $$$$$$$
 \____  $$|__/   \___/ |_______/  \_______/|__/      |__/|_______/  \_______/
 /$$  \ $$                                                                   
|  $$$$$$/                                                                   
 \______/                                                                    `

	colors := []string{
		"#FFFFFF",
		"#A0A0A0",
		"#505050",
	}

	lines := strings.Split(ascii, "\n")

	validLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			validLines++
		}
	}

	lineIndex := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			fmt.Println()
			continue
		}

		colorIndex := 0
		if validLines > 1 {
			colorIndex = int(float64(lineIndex) / float64(validLines-1) * float64(len(colors)-1))
		}
		if colorIndex >= len(colors) {
			colorIndex = len(colors) - 1
		}

		styled := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors[colorIndex])).
			Render(line)

		fmt.Println(styled)
		lineIndex++
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

func SuccessIcon() string {
	return "✓"
}

func ErrorIcon() string {
	return "✗"
}

func WarningIcon() string {
	return "⚠"
}

func InfoIcon() string {
	return "ℹ"
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
