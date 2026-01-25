package cmd

import (
	"log"
	"os"

	"github.com/blang/semver"
	"github.com/creativeprojects/go-selfupdate"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update gitscribe to the latest version",
	Run: func(cmd *cobra.Command, args []string) {
		return update()
	},
}

func init() {
}

func update() error {
	currentVersion, err := semver.Parse(version)
	if err != nil {
		log.Println("Error parsing current version (this may happen in dev mode):", err)
		return nil
	}

	latest, err := CheckForUpdate(currentVersion)
	if err != nil {
		log.Println("Error checking for update:", err)
		return nil
	}

	if latest == nil {
		pterm.Info.Println("Current version is the latest")
		return nil
	}

	pterm.DefaultBox.WithTitle("Update Available: v" + latest.Version.String()).Println(latest.ReleaseNotes)
	pterm.Println()

	confirmed, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultText("Do you want to update?").
		Show()

	if !confirmed {
		log.Println("Update canceled")
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		log.Println("Could not locate executable path")
		return nil
	}

	pterm.Info.Println("Updating binary...")
	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		if os.IsPermission(err) {
			log.Println("Permission denied. Please run the update command with sudo: sudo gs update")
			return nil
		}
		log.Println("Error occurred while updating binary:", err)
		return nil
	}
	log.Println("Successfully updated to version", latest.Version)

	return nil
}
