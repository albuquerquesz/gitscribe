package cmd

import (
	"fmt"
	"os"

	"github.com/albqvictor1508/gitscribe/internal/style"
	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update gitscribe to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		return update()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func update() error {
	currentVersion, err := semver.Parse(version)
	if err != nil {
		style.Error(fmt.Sprintf("Error parsing current version (this may happen in dev mode): %v", err))
		return nil
	}

	latest, err := CheckForUpdate(currentVersion)
	if err != nil {
		style.Error(fmt.Sprintf("Error checking for update: %v", err))
		return nil
	}

	if latest == nil {
		style.Info("Current version is the latest")
		return nil
	}

	style.Box("Update Available: v"+latest.Version.String(), latest.ReleaseNotes)

	if !style.InteractiveConfirm("Do you want to update?") {
		style.Info("Update canceled")
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		style.Error("Could not locate executable path")
		return nil
	}

	style.Info("Updating binary...")
	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		if os.IsPermission(err) {
			style.Error("Permission denied. Please run the update command with sudo: sudo gs update")
			return nil
		}
		style.Error(fmt.Sprintf("Error occurred while updating binary: %v", err))
		return nil
	}
	style.Success(fmt.Sprintf("Successfully updated to version %s", latest.Version))

	return nil
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

func ShowUpdate(v string) {
	currentVersion, err := semver.Parse(v)
	if err != nil {
		return
	}
	latest, err := CheckForUpdate(currentVersion)

	if err != nil || latest == nil {
		return
	}
	style.Box("Update Available", "A new version of gitscribe (v"+latest.Version.String()+") is available! Run 'gs update' to get it.")
}