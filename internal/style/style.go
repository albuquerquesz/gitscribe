package style

import (
	"time"

	"github.com/pterm/pterm"
)

func ConfirmAction(msg string) bool {
	pterm.DefaultBox.WithTitle("Commit Suggestion").Println(msg)
	pterm.Println()

	confirmed, _ := pterm.DefaultInteractiveConfirm.
		Show()

	return confirmed
}

func GetASCIIName() {
	ascii := `
           /$$   /$$                                  /$$ /$$
          |__/  | $$
  /$$$$$$  /$$ /$$$$$$   /$$$$$$$  /$$$$$$$  /$$$$$$  /$$| $$$$$$$   /$$$$$$ 
 /$$__  $$| $$|_  $$_/  /$$_____/ /$$_____/ /$$__  $$| $$| $$__  $$ /$$__  $$ 
| $$  \ $$| $$  | $$   |  $$$$$$ | $$      | $$  \__/| $$| $$  \ $$| $$$$$$$
| $$  | $$| $$  | $$ /$$\____  $$| $$      | $$      | $$| $$  | $$| $$_____/
|  $$$$$$$| $$  |  $$$$//$$$$$$$/|  $$$$$$$| $$      | $$| $$$$$$$/|  $$$$$$$
 \____  $$|__/   \___/ |_______/  \_______/|__/      |__/|_______/  \_______/
 /$$  \ $$
|  $$$$$$/
 \______/
`

	pterm.DefaultBasicText.Println(pterm.FgGreen.Sprint(ascii))
	time.Sleep(time.Second)
}

func Spinner(msg string) *pterm.SpinnerPrinter {
	addSpinner, _ := pterm.DefaultSpinner.WithSequence("|", "/", "-", "\\ ").Start()
	addSpinner.UpdateText(msg)

	return addSpinner
}

func Prompt(label string) (string, error) {
	return pterm.DefaultInteractiveTextInput.WithDefaultText(label).WithMask("*").Show()
}

func StringMask(str string) string {
	mask := "****************"
	length := 8

	if len(str) > length {
		mask = str[:length] + "****************"
	}

	return mask
}
