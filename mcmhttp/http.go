package mcmhttp

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/jlmeeker/mcmanager/auth"
	"github.com/jlmeeker/mcmanager/forms"
	"github.com/jlmeeker/mcmanager/releases"
	"github.com/jlmeeker/mcmanager/server"
	"github.com/jlmeeker/mcmanager/vanilla"
)

// APPTITLE is the name of app displayed in the web UI
var APPTITLE string

/*
func getFileSystem() fs.FS {
	//if gin.Mode() == gin.DebugMode {
	//	return os.DirFS("site")
	//}

	fsys, err := fs.Sub(embededFiles, "site")
	if err != nil {
		panic(err)
	}

	return fsys
}
*/

// Listen starts the Gin web server
func Listen(appTitle, addr string, webfiles *fs.FS) {
	APPTITLE = appTitle

	//webfiles := getFileSystem()
	t, err := template.ParseFS(*webfiles, "*.html")
	if err != nil {
		panic(err)
	}

	staticfiles, err := fs.Sub(*webfiles, "static")
	if err != nil {
		panic(err)
	}

	router := gin.Default()
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.SetHTMLTemplate(t)
	router.StaticFS("/img", http.FS(staticfiles))
	router.StaticFS("/js", http.FS(staticfiles))
	router.NoRoute(notFoundHandler)
	router.GET("/", defaultHandler)
	router.GET("/view/:page", defaultHandler)

	v1 := router.Group("/v1")
	{
		v1.GET("/ping", pingHandler)
		v1.GET("/releases", releasesHandler)
		v1.GET("/servers", serversHandler)
		v1.POST("/create", createHandler)
		v1.POST("/delete/:serverid", deleteHandler)
		v1.POST("/start/:serverid", startHandler)
		v1.POST("/stop/:serverid", stopHandler)
		v1.POST("/login", loginHandler)
		v1.POST("/logout", logoutHandler)
		v1.GET("/news", newsHandler)
	}

	router.Run(addr)
}

// PageData defines data that is passed to HTML templates
type PageData struct {
	Authenticated bool
	PlayerName    string
	AppTitle      string
	Page          string
	Releases      struct {
		Vanilla releases.VersionFile
	}
	Servers map[string]server.ServerWebView
	Status  struct {
		Uptime string
	}
}

func notFoundHandler(c *gin.Context) {
	pd := PageData{
		AppTitle: APPTITLE,
		Page:     "notfound",
	}
	c.HTML(http.StatusNotFound, "index.html", pd)
}

// TODO: sanity-check input
func createHandler(c *gin.Context) {
	var success = http.StatusUnauthorized
	var formData forms.NewServer
	var err error
	var s server.Server

	token, _ := c.Cookie("token")
	playerName, _ := c.Cookie("player")
	if auth.VerifyToken(playerName, token) {
		err = c.Bind(&formData)
		if err == nil {
			port := server.NextAvailablePort()
			s, err = server.NewServer(playerName, formData, port)
			if err == nil {
				err = s.AddOp(playerName, true)
				if err != nil {
					success = http.StatusInternalServerError

					// cleanup
					s.Delete()
				} else {
					success = http.StatusOK
					server.LoadServers()
				}
			} else {
				success = http.StatusInternalServerError
			}
		} else {
			success = http.StatusBadRequest
		}
	} else {
		err = errors.New("must be logged in to create a server")
	}

	var data = gin.H{
		"result": success,
		"page":   formData.Page,
	}
	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

func loginHandler(c *gin.Context) {
	var success = http.StatusUnauthorized
	var formData forms.Login

	var data = gin.H{
		"result": success,
	}

	err := c.Bind(&formData)
	if err == nil {
		token, playerName, err := auth.Auth(formData.Username, formData.Password)
		if err == nil {
			c.SetCookie("token", token, 604800, "/", "", true, true) // 604800 = 1 week
			c.SetCookie("player", playerName, 604800, "/", "", true, true)
			data["success"] = http.StatusOK
			data["playername"] = playerName
			data["token"] = token
			data["page"] = formData.Page
			success = http.StatusOK
		}
	} else {
		data["result"] = http.StatusBadRequest
		err = errors.New("login failed")
	}

	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

func logoutHandler(c *gin.Context) {
	var success = http.StatusUnauthorized
	var err error
	token, _ := c.Cookie("token")
	playerName, _ := c.Cookie("player")
	if auth.VerifyToken(playerName, token) {
		c.SetCookie("token", "", 0, "/", "", true, true)
		c.SetCookie("player", "", 0, "/", "", true, true)
		success = http.StatusOK
	} else {
		err = errors.New("not logged in")
	}

	var data = gin.H{
		"result": success,
	}
	if err != nil {
		data["error"] = err.Error()
	}

	c.JSON(success, data)
}

func deleteHandler(c *gin.Context) {
	var success = http.StatusUnauthorized
	var err error
	serverID := c.Param("serverid")
	token, _ := c.Cookie("token")
	playerName, _ := c.Cookie("player")
	if auth.VerifyToken(playerName, token) {
		if s, ok := server.Servers[serverID]; ok {
			if s.Owner == playerName { // only the owner can delete a server
				err = s.Delete()
				if err == nil {
					success = http.StatusOK
					server.LoadServers()
				} else {
					success = http.StatusInternalServerError
				}
			}
		}
	} else {
		err = errors.New("must be logged in to create a server")
		success = http.StatusUnauthorized
	}

	var data = gin.H{
		"result": success,
	}
	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

func startHandler(c *gin.Context) {
	var success = http.StatusUnauthorized
	var err error
	serverID := c.Param("serverid")
	token, _ := c.Cookie("token")
	playerName, _ := c.Cookie("player")
	if auth.VerifyToken(playerName, token) {
		if s, ok := server.Servers[serverID]; ok {
			err = s.Start()
			if err != nil {
				fmt.Printf("WARNING server start unable to fork: %s\n", err.Error())
			}
			success = http.StatusOK
		} else {
			success = http.StatusNotFound
		}
	} else {
		err = errors.New("must be logged in to create a server")
	}

	var data = gin.H{
		"result": success,
	}
	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

func stopHandler(c *gin.Context) {
	var success = http.StatusUnauthorized
	var err error
	serverID := c.Param("serverid")
	token, _ := c.Cookie("token")
	playerName, _ := c.Cookie("player")
	if auth.VerifyToken(playerName, token) {
		if s, ok := server.Servers[serverID]; ok {
			err = s.Stop(10)
			if err == nil {
				success = http.StatusOK
			} else {
				success = http.StatusInternalServerError
			}
		} else {
			success = http.StatusNotFound
		}
	} else {
		err = errors.New("must be logged in to create a server")
	}

	var data = gin.H{
		"result": success,
	}
	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

func defaultHandler(c *gin.Context) {
	pd := PageData{
		AppTitle: APPTITLE,
	}
	pd.Releases.Vanilla = vanilla.Releases

	token, _ := c.Cookie("token")
	playerName, _ := c.Cookie("player")
	if auth.VerifyToken(playerName, token) {
		pd.Authenticated = true
		pd.PlayerName = playerName
	} else {
		fmt.Printf("unauthenticated request: %s\n%s\n", playerName, token)
	}

	status := http.StatusOK
	page := c.Param("page")
	switch page {
	case "", "home":
		pd.Page = "home"
	case "servers":
		pd.Page = "servers"
		pd.Servers = server.OpServersWebView(playerName)
	case "releases":
		pd.Page = "releases"
	case "status":
		pd.Page = "status"
	default:
		pd.Page = "notfound"
		status = http.StatusNotFound
	}

	c.HTML(status, "index.html", pd)
}

func pingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func releasesHandler(c *gin.Context) {
	err := vanilla.RefreshReleases()
	if err != nil {
		c.JSON(502, err)
		return
	}
	c.JSON(200, vanilla.Releases)
}

func serversHandler(c *gin.Context) {
	var result = make(map[string]server.ServerWebView)
	token, _ := c.Cookie("token")
	playerName, _ := c.Cookie("player")
	if !auth.VerifyToken(playerName, token) {
		// silent fail with empty response
		c.JSON(http.StatusOK, result)
		return
	}

	result = server.OpServersWebView(playerName)
	c.JSON(200, result)
}

func newsHandler(c *gin.Context) {
	c.JSON(200, vanilla.News)
}
