package main

import (
	"embed"
	"errors"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed site
var embededFiles embed.FS

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

func listen(addr string) {
	webfiles := getFileSystem()
	t, err := template.ParseFS(webfiles, "*.html")
	if err != nil {
		panic(err)
	}

	staticfiles, err := fs.Sub(webfiles, "static")
	if err != nil {
		panic(err)
	}

	router := gin.Default()
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
		Vanilla VersionFile
	}
	Servers struct {
		List map[string]ServerWebView
	}
	Status struct {
		Uptime string
	}
	News struct {
		Vanilla []NewsItem
	}
}

func notFoundHandler(c *gin.Context) {
	pd := PageData{
		AppTitle: APPTITLE,
		Page:     "notfound",
	}
	c.HTML(http.StatusNotFound, "index.html", pd)
}

// CreateForm expected form fields for creating a new server
type CreateForm struct {
	Name      string `form:"name"`
	MOTD      string `form:"motd"`
	Flavor    string `form:"flavor"`
	Release   string `form:"release"`
	AutoStart bool   `form:"autostart"`
	StartNow  bool   `form:"startnow"`
}

// TODO: sanity-check input
func createHandler(c *gin.Context) {
	var success = http.StatusUnauthorized
	var formData CreateForm
	var err error
	var s Server

	token, _ := c.Cookie("token")
	playerName, _ := c.Cookie("player")
	if verifyToken(playerName, token) {
		err = c.Bind(&formData)
		if err == nil {
			port := nextAvailablePort()
			s, err = NewServer(playerName, formData, port)
			if err == nil {
				err = s.AddOp(playerName, true)
				if err != nil {
					success = http.StatusInternalServerError

					// cleanup
					s.Delete()
				} else {
					success = http.StatusOK
					loadServers()
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
	}
	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

func loginHandler(c *gin.Context) {
	var success = http.StatusUnauthorized
	var formData struct {
		Username string `form:"username"`
		Password string `form:"password"`
	}

	var data = gin.H{
		"result": success,
	}

	err := c.Bind(&formData)
	if err == nil {
		token, playerName, err := auth(formData.Username, formData.Password)
		if err == nil {
			c.SetCookie("token", token, 604800, "/", "", true, true) // 604800 = 1 week
			c.SetCookie("player", playerName, 604800, "/", "", true, true)
			data["success"] = http.StatusOK
			data["playername"] = playerName
			data["token"] = token
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
	if verifyToken(playerName, token) {
		success = http.StatusOK
		removeToken(playerName)
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
	if verifyToken(playerName, token) {
		if s, ok := Servers[serverID]; ok {
			err = s.Delete()
			if err == nil {
				success = http.StatusOK
				loadServers()
			} else {
				success = http.StatusInternalServerError
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
	if verifyToken(playerName, token) {
		if s, ok := Servers[serverID]; ok {
			s.Start()
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
	if verifyToken(playerName, token) {
		if s, ok := Servers[serverID]; ok {
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

	token, _ := c.Cookie("token")
	playerName, _ := c.Cookie("player")
	if verifyToken(playerName, token) {
		pd.Authenticated = true
		pd.PlayerName = playerName
	}

	status := http.StatusOK
	page := c.Param("page")
	switch page {
	case "", "home":
		pd.Page = "home"
		if len(VanillaNews) >= 6 {
			pd.News.Vanilla = VanillaNews[:10]
		}
	case "servers":
		pd.Page = "servers"
		pd.Servers.List = opServersWebView(playerName)
		pd.Releases.Vanilla = VanillaReleases
	case "releases":
		pd.Page = "releases"
		pd.Releases.Vanilla = VanillaReleases
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
	r, err := getReleases()
	if err != nil {
		c.JSON(502, err)
		return
	}
	c.JSON(200, r)
}

func serversHandler(c *gin.Context) {
	var result = make(map[string]ServerWebView)

	playerName, _ := c.Cookie("player")
	if playerName == "" {
		c.JSON(http.StatusForbidden, result)
		return
	}

	result = opServersWebView(playerName)

	c.JSON(200, result)
}
