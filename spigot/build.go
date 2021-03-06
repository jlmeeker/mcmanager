package spigot

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jlmeeker/mcmanager/storage"
)

// Error: rename .storage/spigot/spigot-21w08b.jar .storage/jars/spigot/21w08b.jar: no such file or directory

// Build will create a spigot build for the version specified
func Build(release string) error {
	var err error
	for err == nil {
		err = downloadBuildTools()
		err = doBuild(release)
		err = moveBuildArtifact(release)
		break
	}

	return err
}

// downloadBuildTools will download the BuildTools.jar file (even if it already exists)
func downloadBuildTools() error {
	var err error
	var sources = make(map[string]string)
	sources["buildtools"] = "https://hub.spigotmc.org/jenkins/job/BuildTools/lastSuccessfulBuild/artifact/target/BuildTools.jar"
	for err == nil {
		for _, srcURL := range sources {
			err = storage.DownloadJar("spigot", "BuildTools", srcURL, true)
		}

		break
	}

	return err
}

func doBuild(release string) error {
	cmd := exec.Command("java", "-jar", filepath.Join("../", "../", storage.JARDIR, "spigot", "BuildTools.jar"), "--rev", release)
	cmd.Dir = filepath.Join(storage.SPIGOTBLDDIR)
	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("ERROR building spigot release: %s", release)
	}
	return err
}

func moveBuildArtifact(release string) error {
	var srcPath = filepath.Join(storage.SPIGOTBLDDIR, "spigot-"+release+".jar")
	var dstPath = filepath.Join(storage.JARDIR, "spigot", release+".jar")
	err := os.Rename(srcPath, dstPath)
	if err != nil {
		err = fmt.Errorf("Failed to build spigot for release %s", release)
	}
	return err
}
