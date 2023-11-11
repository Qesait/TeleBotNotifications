package spotify

import (
	"time"
)

const authUrl = "https://accounts.spotify.com"
const apiUrl = "https://api.spotify.com/"

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
