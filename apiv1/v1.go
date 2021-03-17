package apiv1

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jlmeeker/mcmanager/auth"
	"github.com/jlmeeker/mcmanager/forms"
	"github.com/jlmeeker/mcmanager/paper"
	"github.com/jlmeeker/mcmanager/server"
	"github.com/jlmeeker/mcmanager/storage"
	"github.com/jlmeeker/mcmanager/vanilla"
)

//AuditLogMiddleware middleware
//Allow the call to proceed even if an error is encountered while logging
func AuditLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serverID := c.Param("serverid")
		action := c.Param("action")
		playerName, _ := c.Cookie("player")

		storage.AuditWrite(playerName, action, fmt.Sprintf("%s (%s)", serverID, server.Servers[serverID].Name))
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

func V1Routes(v1 *gin.RouterGroup) {
	// these routes available without authorization
	v1.POST("/login", login)
	v1.GET("/news", news)
	v1.GET("/ping", ping)
	v1.GET("/releases", releases)

	// all routes below this line REQUIRE authentication
	v1.Use(AuthenticateMiddleware())
	v1.POST("/create", createHandler)
	v1.POST("/logout", logout)
	v1.GET("/servers", servers)
	v1.GET("/me", me)

	// all routes below this line REQUIRE at least Op access to the requested server
	rgs := v1.Group("/server")
	rgs.Use(server.AuthorizeMiddleware())
	rgs.Use(AuditLogMiddleware())
	rgs.POST("/:serverid/:action", doAction)
}

func doAction(c *gin.Context) {
	var action = c.Param("action")

	switch action {
	case "ado":
		addOp(c)
	case "adw":
		addWhitelist(c)
	case "bkp":
		backup(c)
	case "day":
		day(c)
	case "sav":
		save(c)
	case "wea":
		clearWeather(c)
	case "del":
		delete(c)
	case "rgn":
		regen(c)
	case "sta":
		start(c)
	case "sto":
		stop(c)
	default:
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
}

// addOp ads an op to a server
func addOp(c *gin.Context) {
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

// addWhitelist adds a player to a server's whitelist
func addWhitelist(c *gin.Context) {
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

// backup runs a backup of a server
func backup(c *gin.Context) {
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

// clearWeather sets the server weather to clear
func clearWeather(c *gin.Context) {
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

// createHandler creates a new server instance
func createHandler(c *gin.Context) {
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

// day sets the server time to day
func day(c *gin.Context) {
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

// delete stops and removes a server... permanently
func delete(c *gin.Context) {
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

// login processes a user login request
func login(c *gin.Context) {
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

// logout processes a user logout request
func logout(c *gin.Context) {
	var success = http.StatusOK
	c.SetCookie("token", "", 0, "/", "", true, true)
	c.SetCookie("player", "", 0, "/", "", true, true)

	var data = gin.H{
		"result": success,
		"error":  "",
	}

	c.JSON(success, data)
}

// me get my preferences
func me(c *gin.Context) {
	playerName, _ := c.Cookie("player")
	c.JSON(http.StatusOK, gin.H{
		"hostname":   server.HOSTNAME,
		"result":     http.StatusOK,
		"error":      "",
		"isLoggedIn": true,
		"playerName": playerName,
	})
}

// news returns the current news items
func news(c *gin.Context) {
	var data = gin.H{
		"error": "",
		"news":  vanilla.News,
	}
	c.JSON(http.StatusOK, data)
}

// ping is a simple liveness check
func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
		"error":   "",
	})
}

// regen runs a backup of a server
func regen(c *gin.Context) {
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

// releases returns a list of current, vanilla releases
func releases(c *gin.Context) {
	var success = http.StatusInternalServerError

	err := vanilla.RefreshReleases()
	if err == nil {
		success = http.StatusOK
	} else {
		log.Printf("error getting releases: %s", err.Error())
		err = fmt.Errorf("Unable to get releases")
	}

	var data = gin.H{
		"result":  success,
		"error":   "",
		"vanilla": vanilla.Releases,
		"paper":   paper.Releases,
	}
	if err != nil {
		data["error"] = err.Error()
	}

	c.JSON(success, data)
}

// save tells a server to save data to disk
func save(c *gin.Context) {
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

// servers returns a list of servers the logged in player is either owner of or op on
func servers(c *gin.Context) {
	var result = gin.H{
		"error":   "",
		"servers": make(map[string]server.WebView),
	}

	playerName, _ := c.Cookie("player")
	result["servers"] = server.ServersWebView(playerName)
	c.JSON(http.StatusOK, result)
}

// start starts a server instance
func start(c *gin.Context) {
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

// stop stops a running instance
func stop(c *gin.Context) {
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
