package mcmhttp

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/jlmeeker/mcmanager/apiv1"
	"github.com/jlmeeker/mcmanager/auth"
	"github.com/jlmeeker/mcmanager/releases"
	"github.com/jlmeeker/mcmanager/vanilla"
)

// APPTITLE is the name of app displayed in the web UI
var APPTITLE string

// HOSTNAME is the name of the host as displayed with the port in the web UI
var HOSTNAME string

// Listen starts the Gin web server
func Listen(appTitle, addr string, webfiles *fs.FS, hostname string) error {
	APPTITLE = appTitle
	HOSTNAME = hostname

	if HOSTNAME == "" {
		hn, err := os.Hostname()
		if err == nil {
			HOSTNAME = hn
		} else {
			HOSTNAME = "localhost"
		}
	}

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
	router.GET("/", viewHandler)
	router.GET("/view/:page", viewHandler)

	v1 := router.Group("/api/v1")
	apiv1.V1Routes(v1)

	return router.Run(addr)
}

// PageData defines data that is passed to HTML templates
type PageData struct {
	Authenticated bool
	PlayerName    string
	AppTitle      string
	Hostname      string
	Page          string
	Releases      struct {
		Flavors []string
		Vanilla releases.VersionFile
	}
	//Servers map[string]server.WebView
	Status struct {
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

func viewHandler(c *gin.Context) {
	pd := PageData{
		AppTitle: APPTITLE,
		Hostname: HOSTNAME,
	}
	pd.Releases.Flavors = releases.FLAVORS
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
		//pd.Servers = server.ServersWebView(playerName)
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
