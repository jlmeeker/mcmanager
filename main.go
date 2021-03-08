package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"time"

	"github.com/jlmeeker/mcmanager/auth"
	"github.com/jlmeeker/mcmanager/mcmhttp"
	"github.com/jlmeeker/mcmanager/server"
	"github.com/jlmeeker/mcmanager/storage"
	"github.com/jlmeeker/mcmanager/vanilla"
)

//go:embed site
var embededFiles embed.FS

// APPTITLE is the name of the application as shows in WebUI
const APPTITLE = "MC Manager"

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

	webfiles, err := fs.Sub(embededFiles, "site")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = storage.Prepare(*flagStorageDir)
	if err != nil {
		fmt.Printf("ERROR making storage dirs: %s", err.Error())
		os.Exit(1)
	}

	err = auth.LoadTokenCache()
	if err != nil {
		fmt.Printf("ERROR loading token cache: %s\n", err.Error())
	}

	go func() {
		var err error
		for {
			err = vanilla.RefreshReleases()
			if err != nil {
				fmt.Println(err.Error())
			}

			err = vanilla.RefreshNews(10)
			if err != nil {
				fmt.Println(err.Error())
			}

			time.Sleep(15 * time.Minute)
		}
	}()

	err = server.LoadServers()
	if err != nil {
		fmt.Println(err.Error())
	}

	// Catch interrupt and ask all servers to save (just in case they get culled when we exit)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		for sig := range c {
			fmt.Printf("Received %s... instructing running instances to save\n", sig.String())
			for _, s := range server.Servers {
				if s.IsRunning() {
					err := s.Save()
					if err != nil {
						fmt.Printf(err.Error())
					}
				}
			}
			os.Exit(0)
		}
	}()

	for _, instance := range server.Servers {
		if instance.AutoStart {
			instance.Start()
		}
	}

	mcmhttp.Listen(APPTITLE, *flagListenAddr, &webfiles)
}
