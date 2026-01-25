package style

import "github.com/pterm/pterm"

func ConfirmAction(msg string) bool {
	pterm.DefaultBox.WithTitle("Commit Suggestion").Println(msg)
	pterm.Println()

	confirmed, _ := pterm.DefaultInteractiveConfirm.
		Show()

	return confirmed
}
