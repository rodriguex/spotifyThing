package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	authorize()

	router.GET("/authorize", func(c *gin.Context) {
		token := accessToken(c.Query("code"))
		if token != "" {
			changeSong(token)
		}
	})
	router.Run("localhost:8080")
}

func authorize() {
	apiUrl := "https://accounts.spotify.com/authorize"

	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("redirect_uri", "http://localhost:8080/authorize")
	params.Add("client_id", "88a3f5e29f55459ea2573eecc9534ec2")
	params.Add("scope", "user-read-playback-state, user-modify-playback-state")

	fullUrl := fmt.Sprintf("%s?%s", apiUrl, params.Encode())
	exec.Command("xdg-open", fullUrl).Start()
}

func accessToken(code string) string {
	apiUrl := "https://accounts.spotify.com/api/token"

	formData := url.Values{}
	formData.Set("code", code)
	formData.Set("redirect_uri", "http://localhost:8080/authorize")
	formData.Set("grant_type", "authorization_code")

	base64 := base64.StdEncoding.EncodeToString([]byte("88a3f5e29f55459ea2573eecc9534ec2:d1ece5241ddc4868ba9aeccd42ed6d66"))

	req, repErr := http.NewRequest("POST", apiUrl, strings.NewReader(formData.Encode()))
	if repErr != nil {
		log.Fatalf("%v", repErr)
		return ""
	}

	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+base64)

	client := &http.Client{}
	resp, respErr := client.Do(req)
	if respErr != nil {
		log.Fatalf("%v", respErr)
		return ""
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	var responseMap map[string]interface{}
	if err := json.Unmarshal(body, &responseMap); err != nil {
		fmt.Println(err)
		return ""
	}

	fmt.Println(string(body))

	token, ok := responseMap["access_token"].(string)
	if !ok {
		fmt.Println("error")
		return ""
	}

	return string(token)
}

func changeSong(token string) {
	apiUrl := "https://api.spotify.com/v1/me/player/play"
	data := map[string]interface{}{
		"uris": []string{"spotify:track:7IPKXYU2rTMnrLW5IZ7ZI5"},
	}
	jsonData, _ := json.Marshal(data)

	req, repErr := http.NewRequest("PUT", apiUrl, bytes.NewBuffer(jsonData))
	if repErr != nil {
		log.Fatalf("%v", repErr)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-type", "application/json")

	client := &http.Client{}
	resp, respErr := client.Do(req)
	if respErr != nil {
		log.Fatalf("%v", respErr)
	}

	defer resp.Body.Close()
}
