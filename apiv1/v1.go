package apiv1

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jlmeeker/mcmanager/auth"
	"github.com/jlmeeker/mcmanager/forms"
	"github.com/jlmeeker/mcmanager/server"
	"github.com/jlmeeker/mcmanager/vanilla"
)

func Backup(c *gin.Context) {
	var success = http.StatusUnauthorized
	var err error
	serverID := c.Param("serverid")
	token, _ := c.Cookie("token")
	playerName, _ := c.Cookie("player")
	if auth.VerifyToken(playerName, token) {
		if s, ok := server.Servers[serverID]; ok {
			err = s.Backup()
			if err != nil {
				err = fmt.Errorf("Backup failed: %s", err.Error())
				success = http.StatusInternalServerError
			} else {
				success = http.StatusOK
			}
		} else {
			success = http.StatusNotFound
		}
	} else {
		err = errors.New("must be logged in")
	}

	var data = gin.H{
		"result": success,
	}
	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

func ClearWeather(c *gin.Context) {
	var success = http.StatusUnauthorized
	var err error
	serverID := c.Param("serverid")
	token, _ := c.Cookie("token")
	playerName, _ := c.Cookie("player")
	if auth.VerifyToken(playerName, token) {
		if s, ok := server.Servers[serverID]; ok {
			err = s.WeatherClear()
			if err != nil {
				err = fmt.Errorf("Unable to change weather: %s", err.Error())
				success = http.StatusInternalServerError
			} else {
				success = http.StatusOK
			}
		} else {
			success = http.StatusNotFound
		}
	} else {
		err = errors.New("must be logged in")
	}

	var data = gin.H{
		"result": success,
	}
	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

func Create(c *gin.Context) {
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

func Day(c *gin.Context) {
	var success = http.StatusUnauthorized
	var err error
	serverID := c.Param("serverid")
	token, _ := c.Cookie("token")
	playerName, _ := c.Cookie("player")
	if auth.VerifyToken(playerName, token) {
		if s, ok := server.Servers[serverID]; ok {
			err = s.Day()
			if err != nil {
				err = fmt.Errorf("Unable to set time to day: %s", err.Error())
				success = http.StatusInternalServerError
			} else {
				success = http.StatusOK
			}
		} else {
			success = http.StatusNotFound
		}
	} else {
		err = errors.New("must be logged in")
	}

	var data = gin.H{
		"result": success,
	}
	if err != nil {
		data["error"] = err.Error()
	}
	c.JSON(success, data)
}

func Delete(c *gin.Context) {
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

func Login(c *gin.Context) {
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

func Logout(c *gin.Context) {
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

func News(c *gin.Context) {
	c.JSON(200, vanilla.News)
}

func Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func Releases(c *gin.Context) {
	err := vanilla.RefreshReleases()
	if err != nil {
		c.JSON(502, err)
		return
	}
	c.JSON(200, vanilla.Releases)
}

func Servers(c *gin.Context) {
	var result = make(map[string]server.WebView)
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

func Start(c *gin.Context) {
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

func Stop(c *gin.Context) {
	var success = http.StatusUnauthorized
	var err error
	serverID := c.Param("serverid")
	token, _ := c.Cookie("token")
	playerName, _ := c.Cookie("player")
	if auth.VerifyToken(playerName, token) {
		if s, ok := server.Servers[serverID]; ok {
			err = s.Stop(2)
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
