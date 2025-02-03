package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/gin-gonic/gin"
)

var token string

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("SpotifyThing")

	input := widget.NewEntry()
	input.SetPlaceHolder("type a song")

	execute := widget.NewButton("Search", func() {
		search(input.Text)
		os.Exit(0)
	})

	myWindow.SetContent(container.NewVBox(
		input,
		execute,
	))

	go func() {
		myWindow.Canvas().Focus(input)
	}()

	myWindow.Resize(fyne.NewSize(300, 80))
	myWindow.ShowAndRun()
}

func authorize() {
	apiUrl := "https://accounts.spotify.com/authorize"

	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("redirect_uri", "http://localhost:8080/authorize")
	params.Add("client_id", "3b352d644e58404e80544474ff992166")
	params.Add("scope", "user-read-playback-state,user-modify-playback-state")

	fullUrl := fmt.Sprintf("%s?%s", apiUrl, params.Encode())
	exec.Command("xdg-open", fullUrl).Start()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		gin.SetMode(gin.ReleaseMode)
		router := gin.New()
		router.Use(gin.Recovery())

		router.GET("/authorize", func(c *gin.Context) {
			code := c.Query("code")
			accessToken(code, "auth")
			wg.Done()
		})
		router.Run("localhost:8080")
	}()

	wg.Wait()
}

func accessToken(value string, action string) {
	apiUrl := "https://accounts.spotify.com/api/token"

	formData := url.Values{}
	if action == "auth" {
		formData.Set("code", value)
		formData.Set("grant_type", "authorization_code")
		formData.Set("redirect_uri", "http://localhost:8080/authorize")
	} else {
		formData.Set("grant_type", "refresh_token")
		formData.Set("refresh_token", value)
	}

	req, repErr := http.NewRequest("POST", apiUrl, strings.NewReader(formData.Encode()))
	if repErr != nil {
		log.Fatalf("Error creating request: %v", repErr)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	clientID := "3b352d644e58404e80544474ff992166"
	clientSecret := "fda4e2cdca1244b2819ca6bd805542f0"
	authHeader := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))

	req.Header.Set("Authorization", "Basic "+authHeader)

	client := &http.Client{}
	resp, respErr := client.Do(req)
	if respErr != nil {
		log.Fatalf("Error sending request: %v", respErr)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	var responseMap map[string]interface{}
	if err := json.Unmarshal(body, &responseMap); err != nil {
		log.Fatalf("Error parsing JSON response: %v", err)
	}

	token, _ = responseMap["access_token"].(string)
	refreshToken, _ := responseMap["refresh_token"].(string)

	if action != "auth" {
		os.Remove(".spotifyThingTopSecret.txt")
		os.Remove(".spotifyThingSecret.txt")
	}

	fileToken, _ := os.Create(".spotifyThingTopSecret.txt")
	defer fileToken.Close()
	fileToken.WriteString(token)

	fileRefreshToken, _ := os.Create(".spotifyThingSecret.txt")
	defer fileRefreshToken.Close()
	fileRefreshToken.WriteString(refreshToken)

	if action != "auth" {
		fmt.Println("REFRESH TOKEN RESULTS:")
		fmt.Println(token)
		fmt.Println(refreshToken)
	}
}

func search(song string) {
	if token == "" {
		file, err := os.Open(".spotifyThingTopSecret.txt")
		if err != nil {
			authorize()
		} else {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			scanner.Scan()
			token = scanner.Text()
		}
	}

	apiUrl := "https://api.spotify.com/v1/search"

	params := url.Values{}
	params.Add("q", "track:"+song)
	params.Add("type", "track")
	params.Add("limit", "1")

	fullUrl := fmt.Sprintf("%s?%s", apiUrl, params.Encode())

	req, repErr := http.NewRequest("GET", fullUrl, nil)
	if repErr != nil {
		log.Fatalf("Error creating request: %v", repErr)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, respErr := client.Do(req)
	if respErr != nil {
		log.Fatalf("Error sending request: %v", respErr)
	}

	if resp.StatusCode == 401 {
		file, _ := os.Open(".spotifyThingSecret.txt")
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Scan()

		refreshToken := scanner.Text()

		fmt.Println("REQUESTING REFRESH TOKEN...")
		accessToken(refreshToken, "refresh")
		fmt.Println("AFTER REFRESH TOKEN REQUEST")
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	json.Unmarshal(body, &data)

	tracks, _ := data["tracks"].(map[string]interface{})
	items, _ := tracks["items"].([]interface{})
	if len(items) > 0 {
		song, _ := items[0].(map[string]interface{})
		album, _ := song["album"].(map[string]interface{})
		albumUri, _ := album["uri"].(string)
		trackNumber := song["track_number"].(float64)

		changeSong(albumUri, int(trackNumber)-1)
	}
}

func changeSong(albumUri string, songIndex int) {
	apiUrl := "https://api.spotify.com/v1/me/player/play"

	type Offset struct {
		Position int `json:"position"`
	}

	type Context struct {
		ContextUri string `json:"context_uri"`
		Offset     Offset `json:"offset"`
	}

	data := Context{
		ContextUri: albumUri,
		Offset: Offset{
			Position: songIndex,
		},
	}
	jsonData, _ := json.MarshalIndent(data, "", " ")

	req, repErr := http.NewRequest("PUT", apiUrl, bytes.NewBuffer(jsonData))
	if repErr != nil {
		log.Fatalf("Error creating request: %v", repErr)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, respErr := client.Do(req)
	if respErr != nil {
		log.Fatalf("Error sending request: %v", respErr)
	}

	defer resp.Body.Close()
}
