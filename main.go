package main

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"encoding/base64"
)

func main() {
	router := gin.Default()

	router.GET("/authorize", func(c *gin.Context) {
		apiUrl := "https://accounts.spotify.com/api/token"

		formData := url.Values{}
		formData.Set("code", c.Query("code"))
		formData.Set("redirect_uri", "http://localhost:8080/authorize")
		formData.Set("grant_type", "authorization_code")

		base64 := base64.StdEncoding.EncodeToString([]byte("88a3f5e29f55459ea2573eecc9534ec2:d1ece5241ddc4868ba9aeccd42ed6d66"))

		req, repErr := http.NewRequest("POST", apiUrl, strings.NewReader(formData.Encode()))
		if repErr != nil {
			log.Fatalf("%v", repErr)
		}

		req.Header.Set("content-type", "application/x-www-form-urlencoded")
		req.Header.Set("Authorization", "Basic "+base64)

		client := &http.Client{}
		resp, respErr := client.Do(req)
		if respErr != nil {
			log.Fatalf("%v", respErr)
			return
		}

		defer resp.Body.Close()

		// if resp.StatusCode == 200 {
		// 	playerReq, playerReqErr := http.NewRequest("GET", "https://api.spotify.com/v1/me/player", nil)
		// 	playerReq.Header.set("Authorization", "Bearer "+resp.Body.)
		// }
	})
	router.Run("localhost:8080")
}
