package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// VanillaReleases is a cached copy of the Vanilla json manifest
var VanillaReleases VersionFile

// JARDIR is where we store our .jar files (under STORAGEDIR)
var JARDIR = "jars"

func makeJarDir() error {
	return os.MkdirAll(filepath.Join(STORAGEDIR, JARDIR), 0750)
}

func copyJar(release, serverdir string) error {
	var srcPath = filepath.Join(STORAGEDIR, JARDIR, release+".jar")
	var dstPath = filepath.Join(serverdir, release+".jar")

	// attempt download first (no-op if it exists)
	err := downloadVanillaVersion(release)
	if err != nil {
		return err
	}

	sh, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer sh.Close()

	dh, err := os.OpenFile(dstPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer dh.Close()

	_, err = io.Copy(dh, sh)
	if err != nil {
		return err
	}

	return nil
}

func jarExists(release string, size int64) bool {
	fstat, err := os.Stat(filepath.Join(STORAGEDIR, JARDIR, release+".jar"))
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	if fstat.Size() == size {
		return true
	}
	return false
}

func downloadJar(release, jarURL string, size int64) error {
	if jarExists(release, size) {
		return nil
	}

	var u, err = url.Parse(jarURL)
	if err != nil {
		return err
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !contentTypeIsCorrect("application/octet-stream", resp.Header.Get("Content-Type")) {
		return fmt.Errorf("Wrong content type: %v", resp.Header.Get("Content-Type"))
	}

	expectedBytes := resp.Header.Get("Content-Length")
	ebInt, err := strconv.ParseInt(expectedBytes, 10, 64)
	if err != nil {
		fmt.Println(err.Error())
	}

	var dstPath = filepath.Join(STORAGEDIR, JARDIR, release+".jar")
	dh, err := os.OpenFile(dstPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer dh.Close()

	nBytes, err := io.Copy(dh, resp.Body)
	if err != nil {
		return err
	}

	if ebInt > 0 && nBytes != int64(ebInt) {
		return fmt.Errorf("Failed to download all of the file (%d of %d)", nBytes, ebInt)
	}

	return nil
}

func contentTypeIsCorrect(wanted, contentType string) bool {
	if strings.ToLower(contentType) == wanted {
		return true
	}

	return false
}

// Version is the (important) fields for each release in the version_manifest.json file
type Version struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	ReleaseTime string `json:"releaseTime"`
}

// VersionFile is the structure of the version_manifest.json file
type VersionFile struct {
	Latest struct {
		Release  string
		Snapshot string
	}
	Versions []Version
}

func getVanillaManifest() (VersionFile, error) {
	var v VersionFile
	var manifestURL = "https://launchermeta.mojang.com/mc/game/version_manifest.json"
	resp, err := http.Get(manifestURL)
	if err != nil {
		return v, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return v, err
	}

	err = json.Unmarshal(body, &v)
	if err != nil {
		return v, err
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
	return v, err
}

func getVanillaVersionDownloadURL(versionURL, version string) (string, int64, error) {
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
		return "", 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &vDetails)
	if err != nil {
		return "", 0, err
	}

	if vDetails.ID == version {
		return vDetails.Downloads.Server.URL, vDetails.Downloads.Server.Size, nil
	}

	return "", 0, nil
}

func downloadLatestVanilla() error {
	manifest, err := getVanillaManifest()
	if err != nil {
		return err
	}
	VanillaReleases = manifest

	for _, v := range VanillaReleases.Versions {
		if v.ID == VanillaReleases.Latest.Release || v.ID == VanillaReleases.Latest.Snapshot {
			vURL, vSIZE, err := getVanillaVersionDownloadURL(v.URL, v.ID)
			if err != nil {
				return err
			}
			err = downloadJar(v.ID, vURL, vSIZE)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func downloadVanillaVersion(version string) error {
	manifest, err := getVanillaManifest()
	if err != nil {
		return err
	}
	VanillaReleases = manifest

	for _, v := range VanillaReleases.Versions {
		if v.ID == version {
			vURL, vSIZE, err := getVanillaVersionDownloadURL(v.URL, v.ID)
			if err != nil {
				return err
			}
			err = downloadJar(v.ID, vURL, vSIZE)
			if err != nil {
				return err
			}
			break
		}
	}

	return nil
}

func getReleases() ([]string, error) {
	var r []string
	entries, err := os.ReadDir(filepath.Join(STORAGEDIR, "jars"))
	if err != nil {
		return r, err
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.Contains(name, "jar") {
			continue
		}
		r = append(r, strings.TrimSuffix(name, ".jar"))
	}

	return r, nil
}
