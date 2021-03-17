package paper

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/jlmeeker/mcmanager/releases"
	"github.com/jlmeeker/mcmanager/storage"
)

// Releases is a cached copy of the list of available Paper builds
var Releases releases.VersionFile

func RefreshReleases() error {
	versions, err := getVersions()
	if err != nil {
		return err
	}

	for _, ver := range versions {
		builds, err := getVersionBuilds(ver)
		if err != nil {
			log.Println(err)
			continue
		}

		bld, verr := getVersionBuild(ver, builds.Builds[len(builds.Builds)-1])
		if verr != nil {
			continue
		}
		Releases.Versions = append(Releases.Versions, bld)
	}

	Releases.Latest.Release = Releases.Versions[len(Releases.Versions)-1].ID

	// format dates

	for ndx, version := range Releases.Versions {
		vt, err := time.Parse("2006-01-02T15:04:05.000Z", version.ReleaseTime)
		if err != nil {
			log.Printf("paper time format error: %s", err.Error())
		}
		version.ReleaseTime = vt.Local().Format("2 Jan 2006 3:04 PM")
		Releases.Versions[ndx] = version
	}

	log.Print("paper: finished refreshing releases")
	return nil
}

type PaperProject struct {
	Versions []string `json:"versions"`
}

func getVersions() ([]string, error) {
	var v PaperProject
	var buildsURL = "https://papermc.io/api/v2/projects/paper/"
	resp, err := http.Get(buildsURL)
	if err != nil {
		return v.Versions, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return v.Versions, err
	}

	err = json.Unmarshal(body, &v)
	if err != nil {
		return v.Versions, err
	}

	return v.Versions, err
}

type Builds struct {
	Version string `json:"version"`
	Builds  []int  `json:"builds"`
}

func getVersionBuilds(version string) (Builds, error) {
	var b Builds
	var buildsURL = "https://papermc.io/api/v2/projects/paper/versions/" + version
	resp, err := http.Get(buildsURL)
	if err != nil {
		return b, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return b, err
	}

	err = json.Unmarshal(body, &b)
	if err != nil {
		return b, err
	}

	return b, err
}

type buildInfo struct {
	Build     int    `json:"build"`
	Time      string `json:"time"`
	Version   string `json:"version"`
	Downloads struct {
		Application struct {
			Name string
		} `json:"application"`
	} `json:"downloads"`
}

func getVersionBuild(version string, buildNum int) (releases.Version, error) {
	var v releases.Version
	var build buildInfo
	var err error

	var buildsURL = fmt.Sprintf("https://papermc.io/api/v2/projects/paper/versions/%s/builds/%d", version, buildNum)
	resp, err := http.Get(buildsURL)
	if err != nil {
		return v, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return v, err
	}

	err = json.Unmarshal(body, &build)
	if err != nil {
		return v, err
	}

	v.ID = build.Version
	v.Type = "release"
	v.URL = fmt.Sprintf("https://papermc.io/api/v2/projects/paper/versions/%s/builds/%d/downloads/%s", version, build.Build, build.Downloads.Application.Name)
	v.ReleaseTime = build.Time
	v.Build = buildNum

	return v, err
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
				err = storage.DownloadJar("paper", v.ID, v.URL, true)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
