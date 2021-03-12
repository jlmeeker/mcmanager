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

// HOSTNAME is where we store the flag value
var HOSTNAME = "localhost"

// Hostname takes the flag value and calculates the best attempt at a hostname
func Hostname(flagValue string) {
	var hn string
	if flagValue != "" {
		hn = flagValue
	} else {
		goHostName, err := os.Hostname()
		if err == nil {
			hn = goHostName
		}
	}

	HOSTNAME = hn
}

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
func NewServer(owner string, formData forms.NewServer, port int) (Server, error) {
	var s = Server{
		Name:      formData.Name,
		Owner:     owner,
		Flavor:    formData.Flavor,
		Release:   formData.Release,
		AutoStart: formData.AutoStart,
	}

	var err error
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
		err = s.RefreshProperties()
		s.Props.set("enable-rcon", "true")
		s.Props.set("gamemode", formData.GameMode)
		s.Props.set("rcon.password", "admin")
		s.Props.set("motd", formData.MOTD)
		s.Props.setPort(port)
		s.Props.set("level-type", formData.WorldType)

		if formData.Seed != "" {
			s.Props.set("level-seed", formData.Seed)
		}

		if formData.Whitelist {
			s.Props.enableWhiteList()
		}

		if formData.Hardcore {
			s.Props.set("hardcore", "true")
		}

		if !formData.PVP {
			s.Props.set("pvp", "false")
		}

		err = s.SaveProps()
		err = acceptEULA(s.ServerDir())
		err = storage.DeployJar(s.Flavor, s.Release, s.UUID)
		err = s.SaveManagedJSON()
		pUUID, err = auth.PlayerUUIDLookup(owner)
		err = s.AddOpOffline(owner, pUUID, true)

		if s.WhitelistEnabled() {
			err = s.WhitelistAddOffline(owner, pUUID, true)
		}

		err = storage.SetupServerBackup(s.UUID)
		err = storage.AuditWrite(s.Owner, "create", fmt.Sprintf("created server %s", s.UUID))
		break
	}

	if err == nil && formData.StartNow {
		err = s.Start()
	}

	return s, err
}

// loadManagedJSON reads in the managed.json file
// Does NOT read in server.properties (will likely be stale)
func loadManagedJSON(serverDir string) (Server, error) {
	var s Server
	var err error
	b, err := os.ReadFile(filepath.Join(serverDir, "managed.json"))
	if err != nil {
		return s, err
	}

	err = json.Unmarshal(b, &s)
	return s, err
}

// LoadServer creates a new instance of Server from an existing serverdir
func LoadServer(serverDir string) (Server, error) {
	var s Server
	var err error

	for err == nil {
		s, err = loadManagedJSON(serverDir)
		err = s.RefreshProperties()
		break
	}

	// Save here to get new properties written to managed.json
	return s, s.SaveManagedJSON()
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
func (s *Server) AddOpOffline(opName, uuid string, force bool) error {
	ops, err := s.LoadOps()
	if err != nil && force == false {
		return err
	}

	var o = Op{
		UUID:              uuid,
		Name:              opName,
		Level:             4,
		BypassPlayerLimit: true,
	}

	ops = append(ops, o)
	storage.AuditWrite("server_AddOpOffline", "create", fmt.Sprintf("opped %s on %s", opName, s.UUID))
	return s.SaveOps(ops)
}

// AddOpOnline will add a user as an op using rcon
func (s *Server) AddOpOnline(opName string) error {
	if opName == "" {
		return fmt.Errorf("cannot op that user")
	}

	var err error
	for err == nil {
		_, err = s.rcon(fmt.Sprintf("op %s", opName))
		err = storage.AuditWrite("server_AddOpOnline", "op:add", fmt.Sprintf("opped %s on %s", opName, s.UUID))

		if s.WhitelistEnabled() {
			err = s.WhitelistAddOnline(opName)
		}

		break
	}

	return err
}

// Backup will instruct the server to perform a save-all operation
func (s *Server) Backup(message string) error {
	return storage.GitCommit(s.UUID, message)
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

// LoadWhitelist will read in the contents of the server's ops.json file
func (s *Server) LoadWhitelist() ([]WLPlayer, error) {
	var wlps []WLPlayer
	b, err := os.ReadFile(filepath.Join(s.ServerDir(), "whitelist.json"))
	if err != nil {
		return wlps, err
	}

	err = json.Unmarshal(b, &wlps)
	if err != nil {
		return wlps, err
	}
	return wlps, nil
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

// RefreshProperties reads in the server.properties values
func (s *Server) RefreshProperties() error {
	p, err := loadProperties(s.ServerDir())
	if err == nil {
		s.Props = p
	}
	return err
}

// Regen generates a new world
func (s *Server) Regen() error {
	var err error
	var running = s.IsRunning()
	for err == nil {
		if running {
			err = s.Stop(0)
		}

		err = storage.EraseServerFile(s.UUID, "logs")
		err = storage.EraseServerFile(s.UUID, "world")
		err = storage.EraseServerFile(s.UUID, "world_nether")
		err = storage.EraseServerFile(s.UUID, "world_the_end")
		err = storage.EraseServerFile(s.UUID, ".git")
		err = storage.SetupServerBackup(s.UUID)
		err = LoadServers()

		if running {
			err = s.Start()
		}
		break
	}

	return err
}

// Rcon sends a message to the server's rcon
func (s *Server) rcon(msg string) (string, error) {
	//fmt.Printf("server send rcon: %s\n", msg)
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

// SaveProps will save the properties into server.properties
func (s *Server) SaveProps() error {
	return s.Props.writeToFile(s.ServerDir())
}

// SaveWhitelist will save the provided ops to the server's ops.json (overwrites the contents)
func (s *Server) SaveWhitelist(wlps []WLPlayer) error {
	b, err := json.MarshalIndent(wlps, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(s.ServerDir(), "whitelist.json"), b, 0640)
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

	var err error
	for err == nil {
		_, err = s.rcon(fmt.Sprintf("/say Server shutting down in %d seconds", delay))
		time.Sleep(time.Duration(delay) * time.Second)
		_, err = s.rcon("stop")
		break
	}

	if err == nil {
		for s.IsRunning() {
			time.Sleep(1 * time.Second)
		}
	}
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

// WhitelistAddOnline will instruct the server to whitelist a player
func (s *Server) WhitelistAddOnline(playerName string) error {
	_, err := s.rcon(fmt.Sprintf("whitelist add %s", playerName))
	if err != nil {
		return err
	}
	err = storage.AuditWrite("server_WhitelistAddOnline", "whitelist:add", fmt.Sprintf("whitelisted %s on %s", playerName, s.UUID))
	return err
}

// WhitelistAddOffline will instruct the server to whitelist a player
func (s *Server) WhitelistAddOffline(playerName, uuid string, force bool) error {
	wlps, err := s.LoadWhitelist()
	if err != nil && force == false {
		return err
	}

	var p = WLPlayer{
		UUID: uuid,
		Name: playerName,
	}

	wlps = append(wlps, p)
	err = storage.AuditWrite("server_WhitelistAddOffline", "create", fmt.Sprintf("whitelisted %s on %s", playerName, s.UUID))
	return s.SaveWhitelist(wlps)
}

// WhitelistEnabled will instruct the server to whitelist a player
func (s *Server) WhitelistEnabled() bool {
	if s.Props.get("white-list") == "true" {
		return true
	}
	return false
}

// WebView web view of a server instance
type WebView struct {
	AmOwner          bool   `json:"amowner"`
	AutoStart        bool   `json:"autostart"`
	Flavor           string `json:"flavor"`
	GameMode         string `json:"gamemode"`
	Hardcore         string `json:"hardcore"`
	MOTD             string `json:"motd"`
	Name             string `json:"name"`
	Ops              string `json:"ops"`
	Owner            string `json:"owner"`
	Players          string `json:"players"`
	Port             string `json:"port"`
	PVP              string `json:"pvp"`
	Release          string `json:"release"`
	Running          bool   `json:"running"`
	Seed             string `json:"seed"`
	UUID             string `json:"uuid"`
	WhiteList        string `json:"whitelist"`
	WhiteListEnabled bool   `json:"whitelistenabled"`
	WorldType        string `json:"worldtype"`
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

		s.RefreshProperties()
		result[s.Name] = WebView{
			AmOwner:          amowner,
			AutoStart:        s.AutoStart,
			Flavor:           s.Flavor,
			GameMode:         s.Props.get("gamemode"),
			Hardcore:         s.Props.get("hardcore"),
			MOTD:             s.Props.get("motd"),
			Name:             s.Name,
			Ops:              strings.Join(ops, ", "),
			Owner:            s.Owner,
			Players:          s.Players(),
			PVP:              s.Props.get("pvp"),
			Port:             s.Props.get("server-port"),
			Release:          s.Release,
			Running:          s.IsRunning(),
			Seed:             s.Props.get("level-seed"),
			UUID:             s.UUID,
			WhiteListEnabled: s.WhitelistEnabled(),
			WhiteList:        s.Whitelist(),
			WorldType:        s.Props.get("level-type"),
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
