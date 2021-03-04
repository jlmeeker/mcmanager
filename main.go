package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

// APPTITLE is the name of the application as shows in WebUI
const APPTITLE = "MC Manager"

// STORAGEDIR is where all server instances are stored
var STORAGEDIR string

// Servers is the global list of managed servers
var Servers = make(map[string]Server)

// Flags
var (
	flagStorageDir = flag.String("storage", "", "where to store server data")
	flagListenAddr = flag.String("listen", "127.0.0.1:8080", "address to listen for http traffic")
)

func main() {
	flag.Parse()

	if *flagStorageDir == "" {
		fmt.Println("option -storage is required")
		os.Exit(1)
	}

	STORAGEDIR = *flagStorageDir
	err := makeStorageDir()
	if err != nil {
		fmt.Printf("ERROR making storage dir: %s", err.Error())
		os.Exit(1)
	}

	err = makeJarDir()
	if err != nil {
		fmt.Printf("ERROR making jar dir: %s", err.Error())
		os.Exit(1)
	}

	err = loadTokenCache()
	if err != nil {
		fmt.Printf("ERROR loading token cache: %s\n", err.Error())
	}

	go func() {
		var err error
		for {
			downloadLatestVanilla()
			VanillaNews, _ = vanillaNews(25)
			VanillaReleases, _ = getVanillaManifest()
			if err != nil {
				fmt.Println(err.Error())
			}

			time.Sleep(15 * time.Minute)
		}
	}()

	err = loadServers()
	if err != nil {
		fmt.Println(err.Error())
	}

	// Catch interrupt and ask all servers to backup (just in case they get culled when we exit)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		for sig := range c {
			fmt.Printf("Received %s... backing up all running servers\n", sig.String())
			for _, s := range Servers {
				if s.IsRunning() {
					err := s.Backup()
					if err != nil {
						fmt.Printf(err.Error())
					}
				}
			}
			os.Exit(0)
		}
	}()

	for _, instance := range Servers {
		if instance.AutoStart {
			instance.Start()
		}
	}

	listen(*flagListenAddr)
}

func inList(needle string, haystack []string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}

func loadServers() error {
	var servers = make(map[string]Server)
	var basedir = filepath.Join(STORAGEDIR, "servers")
	entries, err := os.ReadDir(basedir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			var name = entry.Name()
			var entrydir = filepath.Join(basedir, name)
			s, err := LoadServer(entrydir)
			if err != nil {
				fmt.Printf("error loading %s: %s\n", entrydir, err.Error())
			} else {
				servers[s.Name] = s
			}
		}
	}

	Servers = servers
	return nil
}

func makeStorageDir() error {
	return os.MkdirAll(filepath.Join(STORAGEDIR, "servers"), 0750)
}
