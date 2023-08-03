package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	// "bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
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
// type OAuth2TokenParameters struct {
// 	GrantType string `json:"grant_type"`
// 	Code string `json:"code"`
// 	RedirectUri string `json:"redirect_uri"`
// }

// var parameters = OAuth2TokenParameters{"authorization_code", *authorization_code, *redirect_uri}
// parameters_json, err := json.Marshal(parameters)

// request, err := http.NewRequest("POST", url, bytes.NewBuffer(parameters_json))

type OAuth2Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}


func requstAccessToken(client_id *string, client_secret *string, authorization_code *string, redirect_uri *string) OAuth2Token {
	var baseURL = "https://accounts.spotify.com"
	var resource = "/api/token"
	u, _ := url.ParseRequestURI(baseURL)
	u.Path = resource
	urlStr := u.String()

	var parameters = url.Values{}
	parameters.Set("grant_type", "authorization_code")
	parameters.Set("code", *authorization_code)
	parameters.Set("redirect_uri", *redirect_uri)

	request, err := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(parameters.Encode()))
	if err != nil {
		panic(err)
	}
	fmt.Println("Request: ", request)

	var authorization = base64.StdEncoding.EncodeToString([]byte(*client_id + ":" + *client_secret))
	request.Header.Add("Authorization", "Basic "+authorization)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	token := &OAuth2Token{}
	derr := json.NewDecoder(response.Body).Decode(token)
	if derr != nil {
		panic(derr)
	}

	if response.StatusCode != http.StatusOK {
		panic(response.Status)
	}
	
	return *token
}

// b, _ := io.ReadAll(response.Body)
// for key, values := range response.Header {
// 	fmt.Println(key, ":", values)
// }
// fmt.Println(string(b))
func main() {
	const serverPort = 8888
	var c = make(chan string, 1)
	var authorization_code = ""
	var client_id = os.Getenv("spotify_client_id")
	var client_secret = os.Getenv("spotify_client_id")
	var scope = "user-follow-read"
	var redirect_uri = "http://localhost:8888"

	if authorization_code == "" {
		go server(serverPort, c)
		time.Sleep(100 * time.Millisecond)

		var auth_url = generateAuthUrl(&client_id, &redirect_uri, &scope)
		fmt.Println(auth_url)
		// requestURL := fmt.Sprintf("http://localhost:%d?code=NApCCg..BkWtQ&state=34fFs29kd09", serverPort)
		// _, err := http.Get(requestURL)
		// if err != nil {
		// 	fmt.Printf("error making http request: %s\n", err)
		// 	os.Exit(1)
		// }

		authorization_code = <-c // not access token
		fmt.Println(authorization_code)
	}

	fmt.Println(client_id, client_secret, redirect_uri)
	token := requstAccessToken(&client_id, &client_secret, &authorization_code, &redirect_uri)
	fmt.Println("Access token: ", token.AccessToken)
	fmt.Println("Token type: ", token.TokenType)
	fmt.Println("Scope: ", token.Scope)
	fmt.Println("Expires in: ", token.ExpiresIn)
	fmt.Println("Refresh token: ", token.RefreshToken)

}

// var client_id = os.Getenv("spotify_client_id")
// var scope = "user-follow-read"
// var redirect_uri = "http://example.com"

// var auth_url = generateAuthUrl(&client_id, &redirect_uri, &scope)
// fmt.Println(auth_url)
