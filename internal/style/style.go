package style

import (
	"fmt"
	"strings"

	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	"github.com/albuquerquesz/gitscribe/internal/catalog"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	White      = lipgloss.Color("#FFFFFF")
	LightGrey  = lipgloss.Color("#E8E8E8")
	Grey       = lipgloss.Color("#A0A0A0")
	DarkGrey   = lipgloss.Color("#505050")
	SoftRed    = lipgloss.Color("#FF9999")
	SoftOrange = lipgloss.Color("#FFCC99")

	SuccessStyle = lipgloss.NewStyle().Foreground(Grey)
	ErrorStyle   = lipgloss.NewStyle().Foreground(SoftRed)
	InfoStyle    = lipgloss.NewStyle().Foreground(LightGrey)
	WarningStyle = lipgloss.NewStyle().Foreground(SoftOrange)
	TitleStyle   = lipgloss.NewStyle().Foreground(White).Bold(true).MarginBottom(1)

	ProviderHeaderStyle = lipgloss.NewStyle().Foreground(Grey).Bold(true).PaddingLeft(1)
	ModelItemStyle      = lipgloss.NewStyle().PaddingLeft(4)
	DimStyle            = lipgloss.NewStyle().Foreground(Grey)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Grey).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)
)

func GetTheme() *huh.Theme {
	t := huh.ThemeCharm()

	t.Focused.Title = t.Focused.Title.Foreground(White)

	t.Focused.Description = t.Focused.Description.Foreground(Grey)

	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(Grey)

	t.Focused.Option = t.Focused.Option.Foreground(LightGrey)

	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(White)

	t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(Grey)

	t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(Grey)

	t.Blurred.Title = t.Blurred.Title.Foreground(Grey)
	t.Blurred.Option = t.Blurred.Option.Foreground(DarkGrey)

	t.Help.Ellipsis = t.Help.Ellipsis.Foreground(Grey)
	t.Help.ShortKey = t.Help.ShortKey.Foreground(Grey)
	t.Help.ShortDesc = t.Help.ShortDesc.Foreground(LightGrey)
	t.Help.FullKey = t.Help.FullKey.Foreground(Grey)
	t.Help.FullDesc = t.Help.FullDesc.Foreground(LightGrey)

	return t
}

func ConfirmAction(msg string) bool {
	var confirm bool
	err := huh.NewConfirm().
		Title(msg).
		Affirmative("Yes").
		Negative("No").
		Value(&confirm).
		WithTheme(GetTheme()).
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
	).WithTheme(GetTheme())

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
		WithTheme(GetTheme()).
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

func EditMessage(current string) (string, error) {
	edited := current
	err := huh.NewInput().
		Value(&edited).
		WithTheme(GetTheme()).
		Run()
	if err != nil {
		return "", err
	}
	return edited, nil
}

func ShowCommitPrompt(message string) (action string, finalMessage string) {
	currentMessage := message
	resultAction := ""
	resultMessage := ""
	done := false
	lineIndex := 0

	for !done {
		messageLines := strings.Count(currentMessage, "\n") + 1
		totalLines := messageLines + 3

		if lineIndex > 0 {
			fmt.Printf("\033[%dA", totalLines)
			for range totalLines {
				fmt.Print("\033[2K\n")
			}
			fmt.Printf("\033[%dA", totalLines)
		}
		lineIndex++

		messageStyle := lipgloss.NewStyle().
			Foreground(LightGrey).
			MarginTop(1).
			MarginBottom(1)
		fmt.Println(messageStyle.Render(currentMessage))

		keyStyle := lipgloss.NewStyle().Foreground(White)
		bracketStyle := lipgloss.NewStyle().Foreground(Grey)
		labelStyle := lipgloss.NewStyle().Foreground(Grey)

		shortcuts := fmt.Sprintf("%s%s%s %s  %s%s%s %s  %s%s%s %s",
			bracketStyle.Render("["), keyStyle.Render("E"), bracketStyle.Render("]"), labelStyle.Render("Edit"),
			bracketStyle.Render("["), keyStyle.Render("ESC"), bracketStyle.Render("]"), labelStyle.Render("Cancel"),
			bracketStyle.Render("["), keyStyle.Render("↵"), bracketStyle.Render("]"), labelStyle.Render("Continue"),
		)
		fmt.Println(shortcuts)

		keyboard.Listen(func(key keys.Key) (stop bool, err error) {
			switch key.Code {
			case keys.Escape:
				resultAction = "cancel"
				resultMessage = ""
				return true, nil
			case keys.Enter:
				resultAction = "commit"
				resultMessage = currentMessage
				return true, nil
			case keys.RuneKey:
				if key.String() == "e" || key.String() == "E" {
					edited, err := EditMessage(currentMessage)
					if err != nil {
						resultAction = "cancel"
						resultMessage = ""
						return true, nil
					}
					currentMessage = edited
					return true, nil
				}
			}
			return false, nil
		})

		if resultAction != "" {
			done = true
		}
	}

	return resultAction, resultMessage
}
