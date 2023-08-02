package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// const auth_endpoint = "https://accounts.spotify.com/authorize?"

func requestUserAuthorization(client_id *string, redirect_uri *string, scope *string) {

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
    urlStr := fmt.Sprintf("%v", u)
	fmt.Println(urlStr)

    resp, err := http.Get(urlStr)


	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
}

func main() {
	var client_id = os.Getenv("spotify_client_id")
	var scope = "user-follow-read"
	var redirect_uri = "http://example.com"

	requestUserAuthorization(&client_id, &redirect_uri, &scope)

}
