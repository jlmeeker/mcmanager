package apiv1

import (
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
//Allow the call to proceed even if an error is encountered while logging
func AuditLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var action []string

		serverID := c.Param("serverid")
		playerName, _ := c.Cookie("player")

		pathParts := strings.Split(c.Request.URL.Path, "/")
		if len(pathParts) >= 2 {
			action = pathParts[2 : len(pathParts)-1]
		}

		storage.AuditWrite(playerName, strings.Join(action, ":"), fmt.Sprintf("%s (%s)", serverID, server.Servers[serverID].Name))
		c.Next()
	}
}

//AuthenticateMiddleware middleware
//No need to error check the cookie checks, the verify will fail anyway
func AuthenticateMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Cookie("token")
		playerName, _ := c.Cookie("player")

		if auth.VerifyToken(playerName, token) {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"error": "unauthorized",
		})
	}
}

// AddOp ads an op to a server
func AddOp(c *gin.Context) {
	var success = http.StatusInternalServerError
	var formData forms.AddOp

	serverID := c.Param("serverid")
	s := server.Servers[serverID]
	if err := c.Bind(&formData); err != nil {
		return
	}

	err := s.AddOpOnline(formData.OpName)
	if err == nil {
		success = http.StatusOK
	} else {
		log.Printf("add op error: %s", err.Error())
		err = fmt.Errorf("ad op failed")
	}

	var data = gin.H{
		"result": success,
		"error":  "",
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

	serverID := c.Param("serverid")
	if err := c.Bind(&formData); err != nil {
		return
	}

	s := server.Servers[serverID]
	err := s.AddWhitelistOnline(formData.PlayerName)
	if err == nil {
		success = http.StatusOK
	} else {
		log.Printf("whitelist error: %s", err.Error())
		err = fmt.Errorf("whitelist failed")
	}

	var data = gin.H{
		"result": success,
		"error":  "",
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
	if err == nil {
		success = http.StatusOK
	} else {
		log.Printf("backup error: %s", err.Error())
		err = fmt.Errorf("Backup failed")
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
	var success = http.StatusInternalServerError
	var err error

	serverID := c.Param("serverid")
	s := server.Servers[serverID]
	err = s.WeatherClear()
	if err == nil {
		success = http.StatusOK
	} else {
		err = fmt.Errorf("Unable to change weather")
		log.Printf("clearWeather error: %s", err.Error())
	}

	var data = gin.H{
		"result": success,
		"error":  "",
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

	playerName, _ := c.Cookie("player")
	if err := c.Bind(&formData); err != nil {
		return
	}

	port := server.NextAvailablePort()
	s, err := server.NewServer(playerName, formData, port)
	if err == nil {
		success = http.StatusOK
		go server.LoadServers()
	} else {
		log.Printf("create error (%s): %s, attempting to clean up\n", s.Name, err.Error())
		err = fmt.Errorf("Failed to create the server")
		s.Delete() // ignore any error here
	}

	var data = gin.H{
		"result": success,
		"page":   formData.Page,
		"error":  "",
	}

	if err != nil {
		data["error"] = err.Error()
	}

	c.JSON(success, data)
}

// Day sets the server time to day
func Day(c *gin.Context) {
	var success = http.StatusInternalServerError
	var err error

	serverID := c.Param("serverid")
	s := server.Servers[serverID]

	err = s.Day()
	if err == nil {
		success = http.StatusOK
	} else {
		log.Printf("Unable to set time to day: %s", err.Error())
		err = fmt.Errorf("Unable to set time to day")
	}

	var data = gin.H{
		"result": success,
		"error":  "",
	}
	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

// Delete stops and removes a server... permanently
func Delete(c *gin.Context) {
	var success = http.StatusInternalServerError
	var err error

	serverID := c.Param("serverid")
	s := server.Servers[serverID]
	err = s.Delete()
	if err == nil {
		success = http.StatusOK
		go server.LoadServers()
	} else {
		log.Printf("delete error: %s", err.Error())
		err = fmt.Errorf("Unable to delete")
	}

	var data = gin.H{
		"result": success,
		"error":  "",
	}
	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

// Login processes a user login request
func Login(c *gin.Context) {
	var success = http.StatusInternalServerError
	var formData forms.Login

	if err := c.Bind(&formData); err != nil {
		log.Printf("bin error: %s", err.Error())
		return
	}

	var data = gin.H{
		"result": success,
		"error":  "",
	}

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
	} else {
		log.Printf("login error: %s", err.Error())
		err = fmt.Errorf("login failed")
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
		"error":  "",
	}

	c.JSON(success, data)
}

// Me get my preferences
func Me(c *gin.Context) {
	playerName, _ := c.Cookie("player")
	c.JSON(http.StatusOK, gin.H{
		"hostname":   server.HOSTNAME,
		"result":     http.StatusOK,
		"error":      "",
		"isLoggedIn": true,
		"playerName": playerName,
	})
}

// News returns the current news items
func News(c *gin.Context) {
	var data = gin.H{
		"error": "",
		"news":  vanilla.News,
	}
	c.JSON(http.StatusOK, data)
}

// Ping is a simple liveness check
func Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
		"error":   "",
	})
}

// Regen runs a backup of a server
func Regen(c *gin.Context) {
	var success = http.StatusInternalServerError

	serverID := c.Param("serverid")
	s := server.Servers[serverID]
	err := s.Regen()
	if err == nil {
		success = http.StatusOK
	} else {
		log.Printf("Regen failed: %s", err.Error())
		err = fmt.Errorf("Regen failed")
	}

	var data = gin.H{
		"result": success,
		"error":  "",
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
	} else {
		log.Printf("error getting releases: %s", err.Error())
		err = fmt.Errorf("Unable to get releases")
	}

	var data = gin.H{
		"result":   success,
		"error":    "",
		"releases": vanilla.Releases,
	}
	if err != nil {
		data["error"] = err.Error()
	}

	c.JSON(success, data)
}

// Save tells a server to save data to disk
func Save(c *gin.Context) {
	var success = http.StatusInternalServerError

	serverID := c.Param("serverid")
	s := server.Servers[serverID]
	err := s.Save()
	if err == nil {
		success = http.StatusOK
	} else {
		err = fmt.Errorf("Save failed: %s", err.Error())
	}

	var data = gin.H{
		"result": success,
		"error":  "",
	}
	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

// Servers returns a list of servers the logged in player is either owner of or op on
func Servers(c *gin.Context) {
	var result = gin.H{
		"error":   "",
		"servers": make(map[string]server.WebView),
	}

	playerName, _ := c.Cookie("player")
	result["servers"] = server.OpServersWebView(playerName)
	c.JSON(http.StatusOK, result)
}

// Start starts a server instance
func Start(c *gin.Context) {
	var success = http.StatusInternalServerError

	serverID := c.Param("serverid")
	s := server.Servers[serverID]
	err := s.Start()
	if err == nil {
		success = http.StatusOK
	} else {
		log.Printf("start error: %s", err.Error())
		err = fmt.Errorf("failed to start")
	}

	var data = gin.H{
		"result": success,
		"error":  "",
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
	} else {
		log.Printf("stop error: %s", err.Error())
		err = fmt.Errorf("failed to stop")
	}

	var data = gin.H{
		"result": success,
		"error":  "",
	}
	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}
