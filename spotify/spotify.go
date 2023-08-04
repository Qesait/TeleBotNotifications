package spotify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const apiUrl = "https://accounts.spotify.com"

type OAuth2Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type Client struct {
	clientId      string
	authorization string
	redirectUri   string
	scope         string
}

func NewClient(client_id string, client_secret string, redirect_uri string, scope string) Client {
	return Client{
		clientId:      client_id,
		authorization: base64.StdEncoding.EncodeToString([]byte(client_id + ":" + client_secret)),
		redirectUri:   redirect_uri,
		scope:         scope,
	}
}

func (c *Client) GenerateAuthUrl() string {
	resource := "/authorize"
	params := url.Values{}
	params.Add("client_id", c.clientId)
	params.Add("response_type", "code")
	params.Add("redirect_uri", c.redirectUri)
	params.Add("scope", c.scope)

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	u.RawQuery = params.Encode()
	return u.String()
}

func (c *Client) RequestAccessToken(authorization_code *string) (*OAuth2Token, error) {
	resource := "/api/token"
	u, err := url.ParseRequestURI(apiUrl)
	if err != nil {
		return nil, err
	}
	u.Path = resource
	urlStr := u.String()

	parameters := url.Values{}
	parameters.Set("grant_type", "authorization_code")
	parameters.Set("code", *authorization_code)
	parameters.Set("redirect_uri", c.redirectUri)

	request, err := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(parameters.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Authorization", "Basic "+c.authorization)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %s", response.Status)
	}

	token := &OAuth2Token{}
	err = json.NewDecoder(response.Body).Decode(token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (c *Client) refreshAccessToken(token *OAuth2Token) error {
	resource := "/api/token"
	u, err := url.ParseRequestURI(apiUrl)
	if err != nil {
		return err
	}
	u.Path = resource
	urlStr := u.String()

	parameters := url.Values{}
	parameters.Set("grant_type", "refresh_token")
	parameters.Set("refresh_token", token.RefreshToken)

	request, err := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(parameters.Encode()))
	if err != nil {
		return err
	}

	request.Header.Add("Authorization", "Basic "+c.authorization)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(token)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %s", response.Status)
	}

	return nil
}