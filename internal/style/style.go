package style

import (
	"fmt"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	InfoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	WarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	TitleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Bold(true).MarginBottom(1)
	BoxStyle     = lipgloss.NewStyle().
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
	if err != nil {
		return false
	}
	return confirm
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

type SimpleSpinner struct {
	message string
}

func (s *SimpleSpinner) Stop() {
	// Simple cleanup
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
