package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
	"TeleBotNotifications/spotify"
)

func server(port uint, c chan string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println(r)
		fmt.Println(r.URL.Query().Get("idk"))
		c <- r.URL.Query().Get("code")
	})
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("error running http server: %s\n", err)
		}
	}
}
const serverPort = 8888
var c = make(chan string, 1)
var authorization_code = ""

func main() {
	var client_id = os.Getenv("spotify_client_id")
	var client_secret = os.Getenv("spotify_client_secret")
	var scope = "user-follow-read"
	var redirect_uri = "http://localhost:8888"

	if client_id == "" || client_secret == "" {
		panic("no client credentials")
	}
	
	spotify_client, _ := spotify.NewClient(client_id, client_secret, redirect_uri, scope)
	
	if authorization_code == "" {
		go server(serverPort, c)
		time.Sleep(100 * time.Millisecond)

		var auth_url, _ = spotify_client.GenerateAuthUrl()
		fmt.Println(*auth_url)

		authorization_code = <-c // not access token
		fmt.Println(authorization_code)
	}

	token, _ := spotify_client.RequestAccessToken(&authorization_code)
	fmt.Println(*token)
	artists, err := spotify_client.GetFollowedArtists(token)
	fmt.Println("")
	fmt.Println(err)
	if err == nil {
		fmt.Println(len(artists))
		for i:=0; i<len(artists); i++ {
			fmt.Println(artists[i])
		}
	}
}