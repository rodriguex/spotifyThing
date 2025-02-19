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
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load("/home/gus/www/spotifyThing/.tuatia")

	myApp := app.New()
	myWindow := myApp.NewWindow("SpotifyThing")

	input := widget.NewEntry()
	input.SetPlaceHolder("type a song")

	input.OnSubmitted = func(value string) {
		input.SetText("loading...")
		input.FocusLost()
		search(value)
		input.SetText("")
		os.Exit(0)
	}

	myWindow.SetContent(container.NewVBox(
		input,
	))

	myWindow.Canvas().Focus(input)
	myWindow.Resize(fyne.NewSize(300, 50))

	myWindow.ShowAndRun()
}

func authorize() string {
	apiUrl := "https://accounts.spotify.com/authorize"

	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("redirect_uri", os.Getenv("REDIRECT_URI"))
	params.Add("client_id", os.Getenv("CLIENT_ID"))
	params.Add("scope", "user-read-playback-state,user-modify-playback-state")

	fullUrl := fmt.Sprintf("%s?%s", apiUrl, params.Encode())
	exec.Command("xdg-open", fullUrl).Start()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	var token string

	go func() {
		gin.SetMode(gin.ReleaseMode)
		router := gin.New()
		router.Use(gin.Recovery())

		router.GET("/authorize", func(c *gin.Context) {
			code := c.Query("code")
			token = accessToken(code, "auth")
			wg.Done()
		})
		router.Run(os.Getenv("SERVER_URL"))
	}()

	wg.Wait()
	return token
}

func accessToken(value string, action string) string {
	apiUrl := "https://accounts.spotify.com/api/token"

	formData := url.Values{}
	if action == "auth" {
		formData.Set("code", value)
		formData.Set("grant_type", "authorization_code")
		formData.Set("redirect_uri", os.Getenv("REDIRECT_URI"))
	} else {
		formData.Set("grant_type", "refresh_token")
		formData.Set("refresh_token", value)
	}

	req, repErr := http.NewRequest("POST", apiUrl, strings.NewReader(formData.Encode()))
	if repErr != nil {
		log.Fatalf("Error creating request: %v", repErr)
	}

	authHeader := base64.StdEncoding.EncodeToString([]byte(os.Getenv("CLIENT_ID") + ":" + os.Getenv("CLIENT_SECRET")))

	req.Header.Set("Authorization", "Basic "+authHeader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	fmt.Println(authHeader)
	return ""

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

	fmt.Println(string(body))
	return ""

	var responseMap map[string]interface{}
	if err := json.Unmarshal(body, &responseMap); err != nil {
		log.Fatalf("Error parsing JSON response: %v", err)
	}

	token, _ := responseMap["access_token"].(string)
	if action == "refresh" {
		os.Remove(".spotifyThingTopSecret.txt")
	}

	fileToken, err := os.Create(".spotifyThingTopSecret.txt")
	defer fileToken.Close()
	fileToken.WriteString(token)

	if action == "auth" {
		refreshToken, _ := responseMap["refresh_token"].(string)
		fileRefreshToken, _ := os.Create(".spotifyThingSecret.txt")
		defer fileRefreshToken.Close()
		fileRefreshToken.WriteString(refreshToken)
	}

	return token
}

func search(songInput string) {
	var token string

	file, err := os.Open(".spotifyThingTopSecret.txt")
	if err != nil {
		token = authorize()
	} else {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		scanner.Scan()
		token = scanner.Text()
	}

	apiUrl := "https://api.spotify.com/v1/search"

	var mode string = ""
	var aux string

	words := strings.Split(songInput, " ")
	for _, word := range words {
		if word == "only" {
			mode = "track"
			aux = "only"
		} else if word == "free" {
			mode = "off"
			aux = "free"
		} else if word == "panas" {
			mode = "panas"
			aux = "panas"
		}
	}

	if mode != "" {
		result := strings.Replace(songInput, aux, "", -1)
		songInput = strings.Join(strings.Fields(result), " ")
	}

	params := url.Values{}
	params.Add("limit", "3")

	if mode == "panas" {
		params.Add("q", songInput)
		params.Add("type", "artist")
	} else {
		params.Add("q", "track:"+songInput)
		params.Add("type", "track")
	}

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
		accessToken(refreshToken, "refresh")

		search(songInput)
		return
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var data map[string]interface{}
	json.Unmarshal(body, &data)

	if mode == "panas" {
		artists, _ := data["artists"].(map[string]interface{})
		items, _ := artists["items"].([]interface{})
		if len(items) > 0 {
			actualArtist, _ := items[0].(map[string]interface{})
			id, _ := actualArtist["id"].(string)
			songs := artistTopTracks(token, id)
			changeSong(token, "", -1, songs, "", "free", true)
		}
	} else {
		tracks, _ := data["tracks"].(map[string]interface{})
		items, _ := tracks["items"].([]interface{})
		if len(items) > 0 {
			song, _ := items[0].(map[string]interface{})
			album, _ := song["album"].(map[string]interface{})
			albumUri, _ := album["uri"].(string)
			trackNumber := song["track_number"].(float64)

			cmd := exec.Command("pgrep", "-x", "spotify")
			output, _ := cmd.Output()
			if len(output) < 1 {
				spotify := exec.Command("spotify")
				spotify.Start()
			}
			changeSong(token, albumUri, int(trackNumber)-1, nil, "", mode, false)
		}
	}
}

func changeSong(token string, albumUri string, songIndex int, songs []string, deviceId string, repeatMode string, panas bool) {
	apiUrl := "https://api.spotify.com/v1/me/player/play"
	var params = url.Values{}

	if deviceId != "" {
		params.Add("device_id", deviceId)
	}

	fullUrl := fmt.Sprintf("%s?%s", apiUrl, params.Encode())

	type Offset struct {
		Position int `json:"position"`
	}

	type Context struct {
		ContextUri string `json:"context_uri"`
		Offset     Offset `json:"offset"`
	}

	type Uris struct {
		Uris []string `json:"uris"`
	}

	var data interface{}

	if !panas {
		data = Context{
			ContextUri: albumUri,
			Offset: Offset{
				Position: songIndex,
			},
		}
	} else {
		data = Uris{
			Uris: songs,
		}
	}
	jsonData, _ := json.MarshalIndent(data, "", " ")

	req, repErr := http.NewRequest("PUT", fullUrl, bytes.NewBuffer(jsonData))
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

	if resp.StatusCode == 404 {
		devicedId := getDeviceId(token)
		changeSong(token, albumUri, songIndex, songs, devicedId, repeatMode, panas)
		return
	}

	if repeatMode != "" {
		setRepeatMode(token, repeatMode)
	}
}

func getDeviceId(token string) string {
	apiUrl := "https://api.spotify.com/v1/me/player/devices"

	req, repErr := http.NewRequest("GET", apiUrl, nil)
	if repErr != nil {
		log.Fatalf("Error creating request: %v", repErr)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, respErr := client.Do(req)
	if respErr != nil {
		log.Fatalf("Error sending request: %v", respErr)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var data map[string]interface{}
	json.Unmarshal(body, &data)

	devices, _ := data["devices"].([]interface{})
	for i := range devices {
		device, _ := devices[i].(map[string]interface{})
		deviceType, _ := device["type"].(string)
		if deviceType == "Computer" {
			return device["id"].(string)
		}
	}

	return ""
}

func setRepeatMode(token string, mode string) {
	apiUrl := "https://api.spotify.com/v1/me/player/repeat"

	params := url.Values{}
	params.Add("state", mode)

	fullUrl := fmt.Sprintf("%s?%s", apiUrl, params.Encode())

	req, repErr := http.NewRequest("PUT", fullUrl, nil)
	if repErr != nil {
		log.Fatalf("Error creating request: %v", repErr)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, respErr := client.Do(req)
	if respErr != nil {
		log.Fatalf("Error sending request: %v", respErr)
	}

	defer resp.Body.Close()
}

func artistTopTracks(token string, artistId string) []string {
	apiUrl := "https://api.spotify.com/v1/artists/" + artistId + "/top-tracks"

	req, repErr := http.NewRequest("GET", apiUrl, nil)
	if repErr != nil {
		log.Fatalf("Error creating request: %v", repErr)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, respErr := client.Do(req)
	if respErr != nil {
		log.Fatalf("Error sending request: %v", respErr)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var data map[string]interface{}
	json.Unmarshal(body, &data)

	songs, _ := data["tracks"].([]interface{})
	var songsUris []string

	if len(songs) > 0 {
		for _, song := range songs {
			actualSong := song.(map[string]interface{})
			id := actualSong["uri"].(string)
			songsUris = append(songsUris, id)
		}
	}
	return songsUris
}

func getPlaybackState(token string) {
	apiUrl := "https://api.spotify.com/v1/me/player"

	req, repErr := http.NewRequest("GET", apiUrl, nil)
	if repErr != nil {
		log.Fatalf("Error creating request: %v", repErr)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, respErr := client.Do(req)
	if respErr != nil {
		log.Fatalf("Error sending request: %v", respErr)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	fmt.Println(string(body))
}
