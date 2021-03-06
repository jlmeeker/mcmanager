package vanilla

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jlmeeker/mcmanager/releases"
	"github.com/jlmeeker/mcmanager/storage"
)

// Releases is a cached copy of the Vanilla json manifest
var Releases releases.VersionFile

// RefreshReleases gets releases
func RefreshReleases() error {
	var v releases.VersionFile
	var manifestURL = "https://launchermeta.mojang.com/mc/game/version_manifest.json"
	resp, err := http.Get(manifestURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &v)
	if err != nil {
		return err
	}

	// format dates
	for ndx, version := range v.Versions {
		vt, err := time.Parse("2006-01-02T15:04:05-07:00", version.ReleaseTime)
		if err != nil {
			fmt.Printf(err.Error())
		}
		version.ReleaseTime = vt.Local().Format("2 Jan 2006 3:04 PM")
		v.Versions[ndx] = version
	}

	Releases = v
	return err
}

func getDownloadURL(versionURL, version string) (string, error) {
	type VersionDetails struct {
		ID        string `json:"id"`
		Downloads struct {
			Server struct {
				Size int64  `json:"size"`
				URL  string `json:"url"`
			}
		} `json:"downloads"`
	}

	var vDetails VersionDetails
	resp, err := http.Get(versionURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &vDetails)
	if err != nil {
		return "", err
	}

	if vDetails.ID == version {
		return vDetails.Downloads.Server.URL, nil
	}

	return "", nil
}

// DownloadReleases will find and download the latest release
// an empty list will download the latest release and the latest snapshot
func DownloadReleases(versions []string) error {
	err := RefreshReleases()
	if err != nil {
		return err
	}

	if len(versions) == 0 {
		versions = []string{Releases.Latest.Release, Releases.Latest.Snapshot}
	}

	for _, v := range Releases.Versions {
		for _, version := range versions {
			if v.ID == version {
				vURL, err := getDownloadURL(v.URL, v.ID)
				if err != nil {
					return err
				}
				err = storage.DownloadJar("vanilla", v.ID, vURL, false)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
