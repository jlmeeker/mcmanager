package apiv1

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jlmeeker/mcmanager/auth"
	"github.com/jlmeeker/mcmanager/forms"
	"github.com/jlmeeker/mcmanager/server"
	"github.com/jlmeeker/mcmanager/storage"
	"github.com/jlmeeker/mcmanager/vanilla"
)

//AuditLogMiddleware middleware
func AuditLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var action []string
		var playerName string
		var err error

		serverID := c.Param("serverid")
		for err == nil {
			playerName, err = c.Cookie("player")
			pathParts := strings.Split(c.Request.URL.Path, "/")
			if len(pathParts) >= 2 {
				action = pathParts[2 : len(pathParts)-1]
			}
			err = storage.AuditWrite(playerName, strings.Join(action, ":"), fmt.Sprintf("%s (%s)", serverID, server.Servers[serverID].Name))
			break
		}

		if err == nil {
			c.Next()
			return
		} else {
			log.Printf("Error with audit log: %s\n", err.Error())
		}

		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

//AuthenticateMiddleware middleware
func AuthenticateMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Cookie("token")
		playerName, _ := c.Cookie("player")

		ok := auth.VerifyToken(playerName, token)
		if ok {
			c.Next()
			return
		}
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

// AddOp ads an op to a server
func AddOp(c *gin.Context) {
	var success = http.StatusInternalServerError
	var formData forms.AddOp
	var err error

	serverID := c.Param("serverid")
	for err == nil {
		err = c.Bind(&formData)
		s := server.Servers[serverID]

		if strings.TrimSpace(formData.OpName) == "" {
			err = fmt.Errorf("invalid player name specified")
		}

		err = s.Backup(fmt.Sprintf("before op %s", formData.OpName))
		err = s.AddOpOnline(formData.OpName)

		if s.WhitelistEnabled() {
			s.WhitelistAddOnline(formData.OpName)
		}

		err = s.Backup(fmt.Sprintf("after op %s", formData.OpName))
		success = http.StatusOK
		break
	}

	var data = gin.H{
		"result": success,
	}

	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

// AddWhitelist adds a player to a server's whitelist
func AddWhitelist(c *gin.Context) {
	var success = http.StatusInternalServerError
	var formData forms.WhitelistAdd
	var err error

	serverID := c.Param("serverid")
	for err == nil {
		err = c.Bind(&formData)

		if strings.TrimSpace(formData.PlayerName) == "" {
			err = fmt.Errorf("invalid player name specified")
		}

		s := server.Servers[serverID]
		err = s.Backup(fmt.Sprintf("before whitelist %s", formData.PlayerName))
		err = s.WhitelistAddOnline(formData.PlayerName)
		err = s.Backup(fmt.Sprintf("after whitelist %s", formData.PlayerName))
		success = http.StatusOK
		break
	}

	var data = gin.H{
		"result": success,
	}

	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

// Backup runs a backup of a server
func Backup(c *gin.Context) {
	var success = http.StatusInternalServerError
	var err error
	serverID := c.Param("serverid")
	s := server.Servers[serverID]
	err = s.Backup("initiated via web")
	if err != nil {
		err = fmt.Errorf("Backup failed: %s", err.Error())
	} else {
		success = http.StatusOK
	}

	var data = gin.H{
		"result": success,
	}
	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

// ClearWeather sets the server weather to clear
func ClearWeather(c *gin.Context) {
	var success = http.StatusUnauthorized
	var err error
	serverID := c.Param("serverid")
	s := server.Servers[serverID]
	err = s.WeatherClear()
	if err != nil {
		err = fmt.Errorf("Unable to change weather: %s", err.Error())
		success = http.StatusInternalServerError
	} else {
		success = http.StatusOK
	}

	var data = gin.H{
		"result": success,
	}
	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

// CreateHandler creates a new server instance
func CreateHandler(c *gin.Context) {
	var success = http.StatusInternalServerError
	var formData forms.NewServer
	var err error
	var s server.Server

	playerName, _ := c.Cookie("player")
	err = c.Bind(&formData)
	for err == nil {
		port := server.NextAvailablePort()
		s, err = server.NewServer(playerName, formData, port)
		break
	}

	var data = gin.H{
		"result": success,
		"page":   formData.Page,
	}

	if err == nil {
		success = http.StatusOK
		server.LoadServers()
	} else {
		data["error"] = err.Error()
		//s.Delete()
		fmt.Printf("error on %s create: %s\n", s.Name, err.Error())
	}

	c.JSON(success, data)
}

// Day sets the server time to day
func Day(c *gin.Context) {
	var success = http.StatusUnauthorized
	var err error
	serverID := c.Param("serverid")
	s := server.Servers[serverID]
	err = s.Day()
	if err != nil {
		err = fmt.Errorf("Unable to set time to day: %s", err.Error())
		success = http.StatusInternalServerError
	} else {
		success = http.StatusOK
	}

	var data = gin.H{
		"result": success,
	}
	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

// Delete stops and removes a server... permanently
func Delete(c *gin.Context) {
	var success = http.StatusUnauthorized
	var err error
	playerName, _ := c.Cookie("player")
	serverID := c.Param("serverid")
	s := server.Servers[serverID]
	if s.Owner == playerName { // only the owner can delete a server
		err = s.Delete()
		if err == nil {
			success = http.StatusOK
			server.LoadServers()
		} else {
			success = http.StatusInternalServerError
		}
	}

	var data = gin.H{
		"result": success,
	}
	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

// Login processes a user login request
func Login(c *gin.Context) {
	var success = http.StatusUnauthorized
	var formData forms.Login

	var data = gin.H{
		"result": success,
	}

	err := c.Bind(&formData)
	if err == nil {
		token, playerName, err := auth.Authenticate(formData.Username, formData.Password)
		if err == nil {
			var secure bool
			if c.Request.Proto == "https" {
				secure = true
			}
			c.SetSameSite(http.SameSiteStrictMode)
			c.SetCookie("token", token, 604800, "/", "", secure, true) // 604800 = 1 week
			c.SetCookie("player", playerName, 604800, "/", "", secure, true)
			data["success"] = http.StatusOK
			data["playername"] = playerName
			data["token"] = token
			data["page"] = formData.Page
			success = http.StatusOK
		}
	} else {
		data["result"] = http.StatusBadRequest
		err = errors.New("bad form data received")
	}

	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

// Logout processes a user logout request
func Logout(c *gin.Context) {
	var success = http.StatusOK
	c.SetCookie("token", "", 0, "/", "", true, true)
	c.SetCookie("player", "", 0, "/", "", true, true)

	var data = gin.H{
		"result": success,
	}

	c.JSON(success, data)
}

// Me get my preferences
func Me(c *gin.Context) {
	playerName, _ := c.Cookie("player")
	c.JSON(http.StatusOK, gin.H{
		"hostname":   server.HOSTNAME,
		"result":     http.StatusOK,
		"isLoggedIn": true,
		"playerName": playerName,
	})
}

// News returns the current news items
func News(c *gin.Context) {
	c.JSON(200, vanilla.News)
}

// Ping is a simple liveness check
func Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

// Regen runs a backup of a server
func Regen(c *gin.Context) {
	var success = http.StatusInternalServerError
	var err error
	serverID := c.Param("serverid")
	s := server.Servers[serverID]
	err = s.Regen()
	if err != nil {
		err = fmt.Errorf("Regen failed: %s", err.Error())
	} else {
		success = http.StatusOK
	}

	var data = gin.H{
		"result": success,
	}
	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

// Releases returns a list of current, vanilla releases
func Releases(c *gin.Context) {
	var success = http.StatusInternalServerError
	err := vanilla.RefreshReleases()
	if err == nil {
		success = http.StatusOK
	}

	c.JSON(success, vanilla.Releases)
}

// Save tells a server to save data to disk
func Save(c *gin.Context) {
	var success = http.StatusInternalServerError
	var err error
	serverID := c.Param("serverid")
	s := server.Servers[serverID]
	err = s.Save()
	if err != nil {
		err = fmt.Errorf("Save failed: %s", err.Error())
	} else {
		success = http.StatusOK
	}

	var data = gin.H{
		"result": success,
	}
	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

// Servers returns a list of servers the logged in player is either owner of or op on
func Servers(c *gin.Context) {
	var result = make(map[string]server.WebView)
	playerName, _ := c.Cookie("player")
	result = server.OpServersWebView(playerName)
	c.JSON(http.StatusOK, result)
}

// Start starts a server instance
func Start(c *gin.Context) {
	var success = http.StatusInternalServerError
	var err error

	serverID := c.Param("serverid")
	for err == nil {
		s := server.Servers[serverID]
		err = s.Backup("before start")
		err = s.Start()
		success = http.StatusOK
		break
	}

	var data = gin.H{
		"result": success,
	}
	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

// Stop stops a running instance
func Stop(c *gin.Context) {
	var success = http.StatusInternalServerError
	var err error
	serverID := c.Param("serverid")
	s := server.Servers[serverID]
	err = s.Stop(2)
	if err == nil {
		success = http.StatusOK
	}

	var data = gin.H{
		"result": success,
	}
	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}
