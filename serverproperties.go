package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ServerProperties is a hash of key:value pairs contained in the server.properties file
type ServerProperties map[string]string

func readServerProperties(serverdir string) (ServerProperties, error) {
	var c = make(ServerProperties)

	fileBytes, err := os.ReadFile(filepath.Join(serverdir, "server.properties"))
	if err != nil {
		return c, err
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(fileBytes))
	for scanner.Scan() {
		var line = scanner.Text()
		if strings.HasPrefix(line, "#") || len(line) < 3 {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			fmt.Printf("error parsing line: '%s'\n", line)
		} else {
			c[parts[0]] = parts[1]
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	return c, nil
}

func (sp ServerProperties) writeToFile(serverdir string) error {
	var keys []string
	for key := range sp {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// open our destination file for writing
	dh, err := os.OpenFile(filepath.Join(serverdir, "server.properties"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer dh.Close()

	for _, ndx := range keys {
		_, err := dh.WriteString(fmt.Sprintf("%s=%s\n", ndx, sp[ndx]))
		if err != nil {
			fmt.Printf("Unable to complete saving properties to file: %s\n", err.Error())
			return err
		}
	}

	return nil
}

func (sp ServerProperties) set(key, value string) {
	sp[key] = value
}

func (sp *ServerProperties) setPort(port int) {
	sp.set("server-port", fmt.Sprintf("%d", port))
	sp.set("query.port", fmt.Sprintf("%d", port))
	sp.set("rcon.port", fmt.Sprintf("%d", port-10000))
}

func (sp ServerProperties) get(key string) string {
	if val, ok := sp[key]; ok {
		return val
	}
	return ""
}

func acceptEULA(serverdir string) error {
	dh, err := os.OpenFile(filepath.Join(serverdir, "eula.txt"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer dh.Close()

	_, err = dh.WriteString("eula=true\n")
	return err
}

func writeDefaultPropertiesFile(serverDir string) error {
	return os.WriteFile(filepath.Join(serverDir, "server.properties"), []byte(DEFAULTSERVERPROPERTIES), 0600)
}

//DEFAULTSERVERPROPERTIES is the default server properties file
var DEFAULTSERVERPROPERTIES = `
allow-flight=false
allow-nether=true
broadcast-console-to-ops=true
broadcast-rcon-to-ops=true
difficulty=easy
enable-command-block=false
enable-jmx-monitoring=false
enable-query=false
enable-rcon=false
enable-status=true
enforce-whitelist=false
entity-broadcast-range-percentage=100
force-gamemode=false
function-permission-level=2
gamemode=survival
generate-structures=true
generator-settings=
hardcore=false
level-name=world
level-seed=
level-type=default
max-build-height=256
max-players=20
max-tick-time=60000
max-world-size=29999984
motd=A Minecraft Server
network-compression-threshold=256
online-mode=true
op-permission-level=4
player-idle-timeout=0
prevent-proxy-connections=false
pvp=true
query.port=25565
rate-limit=0
rcon.password=
rcon.port=25575
resource-pack=
resource-pack-sha1=
server-ip=
server-port=25565
snooper-enabled=true
spawn-animals=true
spawn-monsters=true
spawn-npcs=true
spawn-protection=16
sync-chunk-writes=true
use-native-transport=true
view-distance=10
white-list=false
`
