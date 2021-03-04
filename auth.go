package main

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
)

// TokenCache is an in-memory cache of user tokens
var TokenCache = make(map[string]string)

func auth(username, password string) (string, string, error) {
	//Auth with Minecraft version 1
	yggdrasilClient := &yggdrasil.Client{}
	authResponse, err := yggdrasilClient.Authenticate(username, password, "Minecraft", 1)
	if err != nil {
		fmt.Println("Error: " + fmt.Sprintf("%v", err))
		return "", "", errors.New(err.ErrorMessage)
	}

	TokenCache[authResponse.SelectedProfile.Name] = authResponse.AccessToken
	return authResponse.AccessToken, authResponse.SelectedProfile.Name, saveTokenCache()
}

func verifyToken(name, token string) bool {
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
	return os.WriteFile(filepath.Join(STORAGEDIR, "token_cache.json"), jb, 0600)
}

func loadTokenCache() error {
	fb, err := os.ReadFile(filepath.Join(STORAGEDIR, "token_cache.json"))
	if err != nil {
		return err
	}
	return json.Unmarshal(fb, &TokenCache)
}

func removeToken(player string) {
	delete(TokenCache, player)
	saveTokenCache()
}

// UUIDplayer is the structure of data expected as a result of a mojang profile request
type UUIDplayer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Legacy bool   `json:"legacy"`
	Demo   bool   `json:"demo"`
}

func playerUUIDLookup(player string) (string, error) {
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

	if !contentTypeIsCorrect("application/json", resp.Header.Get("Content-Type")) {
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
