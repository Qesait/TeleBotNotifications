package main

import (
	"fmt"
	// "io"
	"errors"
	"net/http"
	"net/url"
	"os"
	"time"
)

func generateAuthUrl(client_id *string, redirect_uri *string, scope *string) string {
	baseURL := "https://accounts.spotify.com"
	resource := "/authorize"
	params := url.Values{}
	params.Add("client_id", *client_id)
	params.Add("response_type", "code")
	params.Add("redirect_uri", *redirect_uri)
	params.Add("scope", *scope)

	u, _ := url.ParseRequestURI(baseURL)
	u.Path = resource
	u.RawQuery = params.Encode()
	return fmt.Sprintf("%v", u)
}

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

func main() {
	const serverPort = 8888
	var c = make(chan string, 1)

	go server(serverPort, c)
	time.Sleep(100 * time.Millisecond)
	
	var client_id = os.Getenv("spotify_client_id")
	var scope = "user-follow-read"
	var redirect_uri = "http://localhost:8888"
	
	var auth_url = generateAuthUrl(&client_id, &redirect_uri, &scope)
	fmt.Println(auth_url)
	// requestURL := fmt.Sprintf("http://localhost:%d?code=NApCCg..BkWtQ&state=34fFs29kd09", serverPort)
	// _, err := http.Get(requestURL)
	// if err != nil {
	// 	fmt.Printf("error making http request: %s\n", err)
	// 	os.Exit(1)
	// }
	
	authorization_code := <-c // not access token
	fmt.Println(authorization_code)
}

// var client_id = os.Getenv("spotify_client_id")
// var scope = "user-follow-read"
// var redirect_uri = "http://example.com"

// var auth_url = generateAuthUrl(&client_id, &redirect_uri, &scope)
// fmt.Println(auth_url)