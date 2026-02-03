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
          |__/  | $$                                 | $$|__/
  /$$$$$$  /$$ /$$$$$$   /$$$$$$$  /$$$$$$$  /$$$$$$  /$$| $$$$$$$   /$$$$$$ 
 /$$__  $$| $$|_  $$_/  /$$_____/ /$$_____/ /$$__  $$| $$| $$__  $$ /$$__  $$
| $$  \ $$| $$  | $$   |  $$$$$$ | $$      | $$  \__/| $$| $$  \ $$| $$$$$$$$
| $$  | $$| $$  | $$ /$$\____  $$| $$      | $$      | $$| $$  | $$| $$_____/
|  $$$$$$$| $$  |  $$$$//$$$$$$$/|  $$$$$$$| $$      | $$| $$$$$$$/|  $$$$$$$
 \____  $$|__/   \___/ |_______/  \_______/|__/      |__/|_______/  \_______/
 /$$  \ $$
| $$    | $$
| $$$$$$$$/
|_______/                                                                    
`

	// Paleta cinza/preto: Branco → Cinza claro → Cinza → Cinza escuro → Preto
	colors := []string{
		"#FFFFFF", // Branco puro
		"#E8E8E8", // Cinza muito claro
		"#C0C0C0", // Cinza claro
		"#989898", // Cinza médio-claro
		"#707070", // Cinza médio
		"#484848", // Cinza escuro
		"#202020", // Cinza muito escuro
		"#000000", // Preto
	}

	lines := strings.Split(ascii, "\n")

	// Filtrar linhas vazias no início e fim
	var validLines []int
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			validLines = append(validLines, i)
		}
	}

	// Animação: mostrar linha por linha de cima para baixo
	for idx, lineIdx := range validLines {
		line := lines[lineIdx]

		// Calcular cor baseado na posição
		colorProgress := float64(idx) / float64(len(validLines)-1)
		colorIndex := int(colorProgress * float64(len(colors)-1))
		if colorIndex >= len(colors) {
			colorIndex = len(colors) - 1
		}

		c := colors[colorIndex]
		styledLine := lipgloss.NewStyle().
			Foreground(lipgloss.Color(c)).
			Bold(true).
			Render(line)

		fmt.Println(styledLine)

		// Delay para efeito de "escorrer"
		time.Sleep(80 * time.Millisecond)
	}

	time.Sleep(200 * time.Millisecond)
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
