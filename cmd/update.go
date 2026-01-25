package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/blang/semver"
	"github.com/pterm/pterm"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
)

func UpdateCLI(version string) *cobra.Command {
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update gitscribe to the latest version",
		Run: func(cmd *cobra.Command, args []string) {
			currentVersion, err := semver.Parse(version)
			if err != nil {
				log.Println("Error parsing current version (this may happen in dev mode):", err)
				return
			}

			latest, err := CheckForUpdate(currentVersion)
			if err != nil {
				log.Println("Error checking for update:", err)
				return
			}

			if latest == nil {
				pterm.Info.Println("Current version is the latest")
				return
			}

			pterm.DefaultBox.WithTitle("Update Available: v" + latest.Version.String()).Println(latest.ReleaseNotes)
			pterm.Println()

			confirmed, _ := pterm.DefaultInteractiveConfirm.
				WithDefaultText("Do you want to update?").
				Show()

			if !confirmed {
				log.Println("Update canceled")
				return
			}

			exe, err := os.Executable()
			if err != nil {
				log.Println("Could not locate executable path")
				return
			}

			pterm.Info.Println("Updating binary...")
			if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
				if os.IsPermission(err) {
					log.Println("Permission denied. Please run the update command with sudo: sudo gs update")
					return
				}
				log.Println("Error occurred while updating binary:", err)
				return
			}
			log.Println("Successfully updated to version", latest.Version)
		},
	}
	return updateCmd
}

func CheckForUpdate(currentVersion semver.Version) (*selfupdate.Release, error) {
	latest, found, err := selfupdate.DetectLatest("albqvictor1508/gitscribe")
	if err != nil {
		return nil, fmt.Errorf("error occurred while detecting version: %w", err)
	}

	if !found || latest.Version.LTE(currentVersion) {
		return nil, nil
	}

	return latest, nil
}

func ShowUpdate(version string) {
	currentVersion, err := semver.Parse(version)
	if err != nil {
		return
	}
	latest, err := CheckForUpdate(currentVersion)

	if err != nil || latest == nil {
		return
	}
	pterm.DefaultBox.WithTitle("Update Available").Println("A new version of gitscribe (v" + latest.Version.String() + ") is available! Run 'gs update' to get it.")
}
