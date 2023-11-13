package spotify

import (
	"time"
	"fmt"
	"net/http"
)

const authUrl = "https://accounts.spotify.com"
const apiUrl = "https://api.spotify.com"

type Album struct {
	Id          string
	Name        string
	AlbumType   string
	AlbumGroup  string
	Url         string
	Uri         string
	ImageUrl    string
	ReleaseDate time.Time
	Artists     []Artist
}

// TODO: create internal struct full matching spotify's
type Artist struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type errorResponse struct {
	Error struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	} `json:"error"`
}


func printRequestInfo(req *http.Request) {
	fmt.Println("Request Method:", req.Method)
	fmt.Println("Request URL:", req.URL)
	fmt.Println("Request Proto:", req.Proto)
	fmt.Println("Request Header:")
	for key, values := range req.Header {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}
	fmt.Println("Request Body:", req.Body)
}