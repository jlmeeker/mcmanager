package storage

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Necesary storage directories
var (
	STORAGEDIR string
	JARDIR     string
	SERVERDIR  string
)

// Default file and directory permissions
var (
	DEFAULTFILEPERM fs.FileMode = 0640
	DEFAULTDIRPERM  fs.FileMode = 0750
)

// Prepare will create all the base storage directories
func Prepare(sd string) error {
	STORAGEDIR = sd
	JARDIR = filepath.Join(STORAGEDIR, "jars")
	SERVERDIR = filepath.Join(STORAGEDIR, "servers")

	var err error
	for err == nil {
		err = os.MkdirAll(STORAGEDIR, DEFAULTDIRPERM)
		err = makeSubDir(STORAGEDIR, "jars")
		err = makeSubDir(STORAGEDIR, "servers")
		break
	}

	return err
}

// DeployJar will copy (and download if necessary) the specified release jar into the server directory
func DeployJar(flavor, release, serverID string) error {
	var srcPath = filepath.Join(JARDIR, flavor, release+".jar")
	var dstPath = filepath.Join(SERVERDIR, serverID, release+".jar")

	sh, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer sh.Close()

	dh, err := os.OpenFile(dstPath, os.O_RDWR|os.O_CREATE, DEFAULTFILEPERM)
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

// JarExists returns whether or now a jar file is available in storage
func JarExists(flavor, release string) bool {
	_, err := os.Stat(filepath.Join(JARDIR, flavor, release+".jar"))
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	return false
}

// DownloadJar downloads a jar file
func DownloadJar(flavor, release, jarURL string, size int64) error {
	if JarExists(flavor, release) {
		return nil
	}

	// no-op if alredy exists
	err := makeSubDir(JARDIR, flavor)
	if err != nil {
		return err
	}

	u, err := url.Parse(jarURL)
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

	var dstPath = filepath.Join(JARDIR, flavor, release+".jar")
	dh, err := os.OpenFile(dstPath, os.O_RDWR|os.O_CREATE, DEFAULTFILEPERM)
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

// GetReleases returns a list of jar files (minus the extension) for the specified flavor
func GetReleases(flavor string) ([]string, error) {
	var r []string
	entries, err := os.ReadDir(filepath.Join(JARDIR, flavor))
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

// DeleteServer is WAY scary!!!
func DeleteServer(id string) error {
	var path = filepath.Join(SERVERDIR, id)
	return os.RemoveAll(path)
}

// MakeServerDir creates the server directory
func MakeServerDir(id string) error {
	return makeSubDir(SERVERDIR, id)
}

// WriteServerFile will save the given filename (and content) to the server directory
func WriteServerFile(id, fname string, content []byte) error {
	var path = filepath.Join(SERVERDIR, id, fname)
	return os.WriteFile(path, content, DEFAULTFILEPERM)
}

func makeSubDir(parent, dirname string) error {
	var path = filepath.Join(parent, dirname)
	return os.MkdirAll(path, DEFAULTDIRPERM)
}
