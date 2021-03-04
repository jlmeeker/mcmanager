package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// FLAVORS is a list of supported minecraft flavors
var FLAVORS = []string{"vanilla", "spigot"}

// Server is an instance of a server, tracked during runtime
type Server struct {
	Name      string           `json:"name"`
	Owner     string           `json:"owner"`
	Props     ServerProperties `json:"properties"`
	Release   string           `json:"release"`
	MaxMem    string           `json:"maxmem"`
	MinMem    string           `json:"minmem"`
	AutoStart bool             `json:"autostart"`
	Flavor    string           `json:"flavor"`
}

// NewServer creates a new instance of Server, and sets up the serverdir
func NewServer(owner string, formData CreateForm, port int) (Server, error) {
	var s = Server{
		Name:      formData.Name,
		Owner:     owner,
		Flavor:    formData.Flavor,
		Release:   formData.Release,
		AutoStart: formData.AutoStart,
	}

	var err error
	var props ServerProperties
	for err == nil {
		if !inList(s.Flavor, FLAVORS) {
			err = errors.New("invalid flavor")
		}

		if port == 0 {
			err = errors.New("unable to find an available port")
		}

		err = os.Mkdir(s.ServerDir(), 0770)
		if err != nil {
			return s, err
		}

		err = writeDefaultPropertiesFile(s.ServerDir())

		props, err = readServerProperties(s.ServerDir())

		props.set("enable-rcon", "true")
		props.set("rcon.password", "admin")
		props.set("motd", formData.MOTD)
		props.setPort(port)

		err = props.writeToFile(s.ServerDir())
		s.Props = props

		err = acceptEULA(s.ServerDir())
		err = copyJar(s.Release, s.ServerDir())
		break
	}

	err = s.Save()

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
	var props ServerProperties
	for err == nil {
		props, err = readServerProperties(serverDir)
		s.Props = props
		break
	}

	// Save here to get new properties written to managed.json
	return s, s.Save()
}

// Players gets player list
func (s *Server) Players() string {
	reply, err := s.Rcon("list")
	if err != nil {
		return ""
	}

	parts := strings.Split(reply, ":")
	if len(parts) < 2 {
		return ""
	}

	return parts[1]
}

// Start starts the server (expected to be run as a goroutine)
func (s *Server) Start() error {
	if s.IsRunning() {
		return errors.New("server already running")
	}

	var cwd = s.ServerDir()
	var cmd = exec.Command("java", "-Xms512M", "-Xmx2G", "-jar", s.Release+".jar", "--nogui")
	cmd.Dir = cwd
	err := cmd.Start()
	if err != nil {
		return err
	}

	return cmd.Process.Release()
}

// Rcon sends a message to the server's rcon
func (s *Server) Rcon(msg string) (string, error) {
	return rconSend(msg, s.Props["rcon.port"], s.Props["rcon.password"])
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

// Save writes the server config to disk
func (s *Server) Save() error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(s.ServerDir(), "managed.json"), b, 0640)
}

// ServerDir builds the path to the server storage dir
func (s *Server) ServerDir() string {
	return filepath.Join(STORAGEDIR, "servers", s.Name)
}

// Stop broadcasts a message to the server then stops it after the delay
func (s *Server) Stop(delay int) error {
	// Stopping a non-running server is a no-op
	if !s.IsRunning() {
		return nil
	}

	_, err := s.Rcon(fmt.Sprintf("/say Server shutting down in %d seconds", delay))
	if err != nil {
		return err
	}

	time.Sleep(time.Duration(delay) * time.Second)
	_, err = s.Rcon("stop")
	return err
}

// Delete is WAY scary!!!
func (s *Server) Delete() error {
	var err error
	for err == nil {
		if !filepath.HasPrefix(s.ServerDir(), filepath.Join(STORAGEDIR, "servers")) {
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

// Ops will return a list of ops contained in the server's ops.json file (a zero-error equivalent of LoadOps)
func (s *Server) Ops() []Op {
	ops, err := s.LoadOps()
	if err != nil {
		fmt.Printf("ERROR loading ops: %s\n", err.Error())
	}
	return ops
}

// ServerWebView web view of a server instance
type ServerWebView struct {
	Name      string `json:"name"`
	Release   string `json:"release"`
	Running   bool   `json:"running"`
	Port      string `json:"port"`
	AutoStart bool   `json:"autostart"`
	Players   string `json:"players"`
	MOTD      string `json:"motd"`
	Flavor    string `json:"flavor"`
	Ops       string `json:"ops"`
}

func opServersWebView(opName string) map[string]ServerWebView {
	var result = make(map[string]ServerWebView)
	if opName == "" {
		return result
	}
	for _, s := range ServersWithOp(opName) {
		var ops []string
		for _, op := range s.Ops() {
			ops = append(ops, op.Name)
		}
		result[s.Name] = ServerWebView{
			Name:      s.Name,
			Release:   s.Release,
			Running:   s.IsRunning(),
			Port:      s.Props["server-port"],
			AutoStart: s.AutoStart,
			Players:   s.Players(),
			MOTD:      s.Props["motd"],
			Flavor:    s.Flavor,
			Ops:       strings.Join(ops, ", "),
		}
	}

	return result
}

// Op is the structure of an op within the ops.json file
type Op struct {
	UUID              string `json:"uuid"`
	Name              string `json:"name"`
	Level             int8   `json:"level"`
	BypassPlayerLimit bool   `json:"bypassesPlayerLimit"`
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

// SaveOps will save the provided ops to the server's ops.json (overwrites the contents)
func (s *Server) SaveOps(ops []Op) error {
	b, err := json.MarshalIndent(ops, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(s.ServerDir(), "ops.json"), b, 0640)
}

// AddOp will add a user as an op (loading the ops.json file contents first, unless forced)
func (s *Server) AddOp(name string, force bool) error {
	ops, err := s.LoadOps()
	if err != nil && force == false {
		return err
	}

	uuid, err := playerUUIDLookup(name)
	if err != nil {
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

// nextAvailablePort will return the next available server port
// one should loadServers() before this to ensure accurate data
func nextAvailablePort() int {
	var err error
	var highest = 25564 // one less than the default server port (we increment it on return)

	for err == nil {
		// loop over servers and find highest used port number
		err = loadServers() // get current server data

		for _, s := range Servers {
			port := s.Props.get("server-port")
			if port == "" {
				continue
			}
			var portInt int
			portInt, err = strconv.Atoi(port)
			fmt.Printf("str: %s, int: %d\n", port, portInt)
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