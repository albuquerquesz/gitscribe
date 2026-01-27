package version

import (
	"fmt"

	"github.com/albuquerquesz/gitscribe/internal/style"
	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

func CheckForUpdate(currentVersion semver.Version) (*selfupdate.Release, error) {
	latest, found, err := selfupdate.DetectLatest("albuquerquesz/gitscribe")
	if err != nil {
		return nil, fmt.Errorf("error occurred while detecting version: %w", err)
	}

	if !found || latest.Version.LTE(currentVersion) {
		return nil, nil
	}

	return latest, nil
}

func ShowUpdate(v string) {
	currentVersion, err := semver.ParseTolerant(v)
	if err != nil {
		return
	}
	latest, err := CheckForUpdate(currentVersion)

	if err != nil || latest == nil {
		return
	}
	style.Box("Update Available", "A new version of gitscribe (v"+latest.Version.String()+") is available! Run 'gs update' to get it.")
}
