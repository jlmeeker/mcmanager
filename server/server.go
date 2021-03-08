package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jlmeeker/mcmanager/auth"
	"github.com/jlmeeker/mcmanager/forms"
	"github.com/jlmeeker/mcmanager/rcon"
	"github.com/jlmeeker/mcmanager/releases"
	"github.com/jlmeeker/mcmanager/spigot"
	"github.com/jlmeeker/mcmanager/storage"
	"github.com/jlmeeker/mcmanager/vanilla"
)

//AuthorizeOpMiddleware middleware
func AuthorizeOpMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		playerName, _ := c.Cookie("player")
		serverID := c.Param("serverid")
		if s, ok := Servers[serverID]; ok {
			if s.IsOwner(playerName) || s.IsOp(playerName) {
				c.Next()
				return
			}
		}
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

//AuthorizeOwnerMiddleware middleware
func AuthorizeOwnerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		playerName, _ := c.Cookie("player")
		serverID := c.Param("serverid")
		if s, ok := Servers[serverID]; ok {
			if s.IsOwner(playerName) {
				c.Next()
				return
			}
		}
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

// Servers is the global list of managed servers
var Servers = make(map[string]Server)

// Server is an instance of a server, tracked during runtime
type Server struct {
	Name      string     `json:"name"`
	Owner     string     `json:"owner"`
	Props     Properties `json:"properties"`
	Release   string     `json:"release"`
	MaxMem    string     `json:"maxmem"`
	MinMem    string     `json:"minmem"`
	AutoStart bool       `json:"autostart"`
	Flavor    string     `json:"flavor"`
	UUID      string     `json:"uuid"`
}

// NewServer creates a new instance of Server, and sets up the serverdir
func NewServer(owner string, formData forms.NewServer, port int, whitelist bool) (Server, error) {
	var s = Server{
		Name:      formData.Name,
		Owner:     owner,
		Flavor:    formData.Flavor,
		Release:   formData.Release,
		AutoStart: formData.AutoStart,
	}

	var err error
	var props Properties
	var suuid uuid.UUID
	var pUUID string
	for err == nil {
		suuid, err = uuid.NewRandom()
		s.UUID = suuid.String()

		if !releases.FlavorIsValid(s.Flavor) {
			err = errors.New("invalid flavor")
		}

		if port == 0 {
			err = errors.New("unable to find an available port")
		}

		// attempt download first (no-op if it exists)
		if !storage.JarExists(s.Flavor, s.Release) {
			switch s.Flavor {
			case "vanilla":
				err = vanilla.DownloadReleases([]string{s.Release})
			case "spigot":
				err = spigot.Build(s.Release)
			}
		}

		err = storage.MakeServerDir(s.UUID)
		err = writeDefaultPropertiesFile(s.ServerDir())
		props, err = readServerProperties(s.ServerDir())
		props.set("enable-rcon", "true")
		props.set("rcon.password", "admin")
		props.set("motd", formData.MOTD)
		props.setPort(port)

		if whitelist {
			props.enableWhiteList()
		}

		err = props.writeToFile(s.ServerDir())
		s.Props = props

		err = acceptEULA(s.ServerDir())
		err = storage.DeployJar(s.Flavor, s.Release, s.UUID)
		err = s.SaveManagedJSON()
		pUUID, err = auth.PlayerUUIDLookup(owner)
		err = s.AddOpOffline(owner, pUUID, true)
		err = storage.SetupServerBackup(s.UUID)
		break
	}

	if err == nil && formData.StartNow {
		err = s.Start()
	}

	return s, err
}

// NewServerFromFile reads in the server.json file
func NewServerFromFile(serverDir string) (Server, error) {
	var s Server
	b, err := os.ReadFile(filepath.Join(serverDir, "managed.json"))
	if err != nil {
		return s, err
	}

	err = json.Unmarshal(b, &s)
	return s, err
}

// LoadServer creates a new instance of Server from an existing serverdir
func LoadServer(serverDir string) (Server, error) {
	s, err := NewServerFromFile(serverDir)
	var props Properties
	for err == nil {
		props, err = readServerProperties(serverDir)
		s.Props = props
		break
	}

	// Save here to get new properties written to managed.json
	return s, s.Save()
}

// LoadServers loads servers from disk and caches results
func LoadServers() error {
	var servers = make(map[string]Server)
	var basedir = filepath.Join(storage.STORAGEDIR, "servers")
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
				servers[s.UUID] = s
			}
		}
	}

	Servers = servers
	return nil
}

// AddOpOffline will add a user as an op
// the force option is to ignore errors from loading the ops.json file
// like happens when one doesn't exist.
func (s *Server) AddOpOffline(name, uuid string, force bool) error {
	ops, err := s.LoadOps()
	if err != nil && force == false {
		return err
	}

	var o = Op{
		UUID:              uuid,
		Name:              name,
		Level:             4,
		BypassPlayerLimit: true,
	}

	ops = append(ops, o)
	return s.SaveOps(ops)
}

// AddOpOnline will add a user as an op using rcon
func (s *Server) AddOpOnline(playerName string) error {
	if playerName == "" {
		return fmt.Errorf("cannot op that user")
	}
	_, err := s.rcon(fmt.Sprintf("op %s", playerName))
	if err != nil {
		return err
	}
	return nil
}

// Backup will instruct the server to perform a save-all operation
func (s *Server) Backup(message string) error {
	return storage.GitCommit(s.UUID, "message")
}

// Day will instruct the server to set the time to day
func (s *Server) Day() error {
	_, err := s.rcon("time set day")
	if err != nil {
		return err
	}
	return nil
}

// Delete is WAY scary!!!
func (s *Server) Delete() error {
	var err error
	for err == nil {
		if !filepath.HasPrefix(s.ServerDir(), filepath.Join(storage.SERVERDIR)) {
			err = errors.New("refusing to delete " + s.ServerDir())
		}

		// Stop it
		if s.IsRunning() {
			err = s.Stop(0)
		}

		// Make sure it is stopped before removing files
		for s.IsRunning() {
			time.Sleep(1 * time.Second)
		}

		err = os.RemoveAll(s.ServerDir())
		break
	}

	return err
}

// IsOp returns if a given player name is found in the list of server ops
func (s *Server) IsOp(player string) bool {
	var ops = s.Ops()
	for _, op := range ops {
		if op.Name == player {
			return true
		}
	}
	return false
}

// IsOwner returns if a given player name the server owner
func (s *Server) IsOwner(player string) bool {
	if player == s.Owner {
		return true
	}
	return false
}

// IsRunning attempts to determine if the server is running by checking rcon connect
func (s *Server) IsRunning() bool {
	conn, err := net.Dial("tcp", "localhost:"+s.Props["rcon.port"])
	if err == nil {
		conn.Close()
		return true
	}
	return false
}

// LoadOps will read in the contents of the server's ops.json file
func (s *Server) LoadOps() ([]Op, error) {
	var o []Op
	b, err := os.ReadFile(filepath.Join(s.ServerDir(), "ops.json"))
	if err != nil {
		return o, err
	}

	err = json.Unmarshal(b, &o)
	if err != nil {
		return o, err
	}
	return o, nil
}

// Ops will return a list of ops contained in the server's ops.json file (a zero-error equivalent of LoadOps)
func (s *Server) Ops() []Op {
	ops, err := s.LoadOps()
	if err != nil {
		fmt.Printf("ERROR loading ops: %s\n", err.Error())
	}
	return ops
}

// Players gets player list
func (s *Server) Players() string {
	reply, err := s.rcon("list")
	if err != nil {
		return ""
	}

	parts := strings.Split(reply, ":")
	if len(parts) < 2 {
		return ""
	}

	return parts[1]
}

// Rcon sends a message to the server's rcon
func (s *Server) rcon(msg string) (string, error) {
	return rcon.Send(msg, s.Props["rcon.port"], s.Props["rcon.password"])
}

// Save will instruct the server to perform a save-all operation
func (s *Server) Save() error {
	_, err := s.rcon("save-all")
	if err != nil {
		return err
	}
	return nil
}

// SaveManagedJSON writes the server config to disk
func (s *Server) SaveManagedJSON() error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(s.ServerDir(), "managed.json"), b, 0640)
}

// SaveOps will save the provided ops to the server's ops.json (overwrites the contents)
func (s *Server) SaveOps(ops []Op) error {
	b, err := json.MarshalIndent(ops, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(s.ServerDir(), "ops.json"), b, 0640)
}

// ServerDir builds the path to the server storage dir
func (s *Server) ServerDir() string {
	return filepath.Join(storage.SERVERDIR, s.UUID)
}

// Start starts the server (expected to be run as a goroutine)
func (s Server) Start() error {
	if s.IsRunning() {
		return errors.New("server already running")
	}

	var args = []string{"-Xms512M", "-Xmx2G", "-jar", s.Release + ".jar", "--nogui"}
	var cwd = s.ServerDir()
	var cmd = exec.Command("java", args...)
	cmd.Dir = cwd
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}
	err := cmd.Start()
	if err != nil {
		return err
	}

	return cmd.Process.Release()
}

// Stop broadcasts a message to the server then stops it after the delay
func (s *Server) Stop(delay int) error {
	// Stopping a non-running server is a no-op
	if !s.IsRunning() {
		return nil
	}

	_, err := s.rcon(fmt.Sprintf("/say Server shutting down in %d seconds", delay))
	if err != nil {
		return err
	}

	time.Sleep(time.Duration(delay) * time.Second)
	_, err = s.rcon("stop")
	return err
}

// WeatherClear will instruct the server to perform a save-all operation
func (s *Server) WeatherClear() error {
	_, err := s.rcon("weather clear")
	if err != nil {
		return err
	}
	return nil
}

// Whitelist returns the list of whitelisted players
func (s *Server) Whitelist() string {
	reply, err := s.rcon("whitelist list")
	if err != nil {
		return ""
	}

	parts := strings.Split(reply, ":")
	if len(parts) < 2 {
		return ""
	}

	return parts[1]
}

// WhitelistAdd will instruct the server to whitelist a player
func (s *Server) WhitelistAdd(playerName string) error {
	_, err := s.rcon(fmt.Sprintf("whitelist add %s", playerName))
	if err != nil {
		return err
	}
	return nil
}

// WebView web view of a server instance
type WebView struct {
	Name             string `json:"name"`
	Release          string `json:"release"`
	Running          bool   `json:"running"`
	Port             string `json:"port"`
	AutoStart        bool   `json:"autostart"`
	Players          string `json:"players"`
	MOTD             string `json:"motd"`
	Flavor           string `json:"flavor"`
	Ops              string `json:"ops"`
	UUID             string `json:"uuid"`
	Owner            string `json:"owner"`
	AmOwner          bool   `json:"amowner"`
	WhiteListEnabled string `json:"whitelistenabled"`
	WhiteList        string `json:"whitelist"`
}

// OpServersWebView is a web view of a list of servers
func OpServersWebView(opName string) map[string]WebView {
	var result = make(map[string]WebView)
	if opName == "" {
		return result
	}
	for _, s := range ServersWithOp(opName) {
		var ops []string
		for _, op := range s.Ops() {
			ops = append(ops, op.Name)
		}

		var amowner bool
		if opName == s.Owner {
			amowner = true
		}

		result[s.Name] = WebView{
			Name:             s.Name,
			Release:          s.Release,
			Running:          s.IsRunning(),
			Port:             s.Props["server-port"],
			AutoStart:        s.AutoStart,
			Players:          s.Players(),
			MOTD:             s.Props["motd"],
			Flavor:           s.Flavor,
			Ops:              strings.Join(ops, ", "),
			UUID:             s.UUID,
			Owner:            s.Owner,
			AmOwner:          amowner,
			WhiteListEnabled: s.Props["enforce-whitelist"],
			WhiteList:        s.Whitelist(),
		}
	}

	return result
}

//ServersWithOp returns a list of servers the Op owns or is an op on
func ServersWithOp(opName string) map[string]Server {
	var servers = make(map[string]Server)

	for n, s := range Servers {
		ops := s.Ops()
		for _, op := range ops {
			if op.Name == opName || s.Owner == opName {
				servers[n] = s
				break
			}
		}
	}

	return servers
}

// NextAvailablePort will return the next available server port
// one should loadServers() before this to ensure accurate data
func NextAvailablePort() int {
	var err error
	var highest = 25564 // one less than the default server port (we increment it on return)

	for err == nil {
		// loop over servers and find highest used port number
		err = LoadServers() // get current server data

		for _, s := range Servers {
			port := s.Props.get("server-port")
			if port == "" {
				continue
			}
			var portInt int
			portInt, err = strconv.Atoi(port)
			if portInt > highest {
				highest = portInt
			}
		}
		break
	}

	if err != nil {
		fmt.Println(err.Error())
		return 0
	}

	// rcon port is 10k less than the server port
	return highest + 1
}

func inList(needle string, haystack []string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
