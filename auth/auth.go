package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	yggdrasil "github.com/jlmeeker/go-yggdrasil"
	"github.com/jlmeeker/mcmanager/storage"
)

// TokenCache is an in-memory cache of user tokens
var TokenCache = make(map[string]string)

// Auth will authenticate a login request
func Auth(username, password string) (string, string, error) {
	//Auth with Minecraft version 1
	yggdrasilClient := &yggdrasil.Client{}
	authResponse, err := yggdrasilClient.Authenticate(username, password, "Minecraft", 1)
	if err != nil {
		fmt.Println("Error: " + fmt.Sprintf("%v", err))
		return "", "", errors.New(err.ErrorMessage)
	}

	// if we have a cached token, return it.... otherwise save the received one
	pt, err2 := LookupToken(authResponse.SelectedProfile.Name)
	if err2 != nil {
		fmt.Printf("lookuptoken error: %s\n", err2.Error())
		pt = authResponse.AccessToken
		TokenCache[authResponse.SelectedProfile.Name] = pt
	}

	return pt, authResponse.SelectedProfile.Name, saveTokenCache()
}

// VerifyToken checks the playerName and token against the token cache
func VerifyToken(name, token string) bool {
	for playerName, playerToken := range TokenCache {
		if token == playerToken && name == playerName {
			return true
		}
	}
	return false
}

func saveTokenCache() error {
	jb, err := json.MarshalIndent(TokenCache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(storage.STORAGEDIR, "token_cache.json"), jb, 0600)
}

// LoadTokenCache reads the file contents into memory (why?)
func LoadTokenCache() error {
	fb, err := os.ReadFile(filepath.Join(storage.STORAGEDIR, "token_cache.json"))
	if err != nil {
		return err
	}
	return json.Unmarshal(fb, &TokenCache)
}

// RemoveToken will erase a name/token pair from the in-memory copy of the token cache
func RemoveToken(player string) {
	delete(TokenCache, player)
	saveTokenCache()
}

// LookupToken looks up a token from in-memory token cache
func LookupToken(player string) (string, error) {
	err := LoadTokenCache()
	if err != nil {
		return "", err
	}

	for cachedPlayer, cachedToken := range TokenCache {
		if cachedPlayer == player {
			return cachedToken, nil
		}
	}

	return "", errors.New("player not in cache")
}

// UUIDplayer is the structure of data expected as a result of a mojang profile request
type UUIDplayer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Legacy bool   `json:"legacy"`
	Demo   bool   `json:"demo"`
}

// PlayerUUIDLookup looks up a UUID for a playerName
func PlayerUUIDLookup(player string) (string, error) {
	var reply []UUIDplayer

	data, err := json.Marshal([]string{
		player,
	})
	if err != nil {
		return "", err
	}

	resp, err := http.Post("https://api.mojang.com/profiles/minecraft", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.Header.Get("Content-Type") != "application/json" {
		return "", fmt.Errorf("Wrong content type: %v", resp.Header.Get("Content-Type"))
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	err = json.Unmarshal(body, &reply)
	if err != nil && len(reply) > 0 {
		return "", err
	}

	return fmt.Sprintf("%s-%s-%s-%s-%s", reply[0].ID[0:8], reply[0].ID[8:12], reply[0].ID[12:16], reply[0].ID[16:20], reply[0].ID[20:]), err
}

//{"error":"BadRequestException","errorMessage":" is invalid"}
