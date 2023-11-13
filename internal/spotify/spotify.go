package spotify

import (
	"fmt"
	"net/http"
	"time"
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

type ExternalUrls struct {
	Spotify string `json:"spotify"`
}

type SimplifiedArtist struct {
	ExternalUrls ExternalUrls `json:"external_urls"`
	Href         string       `json:"href"`
	Id           string       `json:"id"`
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	Uri          string       `json:"uri"`
}

type SimplifiedTrack struct {
	Artists          []SimplifiedArtist
	// AvailableMarkets []string     `json:"available_markets"`
	DiscNumber       int          `json:"disc_number"`
	DurationMs       int          `json:"duration_ms"`
	Explicit         bool         `json:"explicit"`
	ExternalUrls     ExternalUrls `json:"external_urls"`
	Href             string       `json:"href"`
	Id               string       `json:"id"`
	IsPlayable       bool         `json:"is_playable"`
	// LinkedFrom
	// Restrictions
	Name         string       `json:"name"`
	TrackNumber int `json:"track_number"`
	Type         string       `json:"type"`
	Uri          string       `json:"uri"`
	// IsLocal
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
