package spotify

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	// "io"
)

const authUrl = "https://accounts.spotify.com"
const apiUrl = "https://api.spotify.com/"

type oAuth2TokenResponce struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}
type OAuth2Token struct {
	AccessToken  string
	TokenType    string
	Scope        string
	Expires      time.Time
	RefreshToken string
}

func (t *OAuth2Token) Expired() bool {
	return time.Now().After(t.Expires)
}

type Client struct {
	client        *http.Client
	clientId      string
	authorization string
	redirectUri   string
	scope         string
}

func NewClient(client_id string, client_secret string, redirect_uri string, scope string) (*Client, error) {
	if client_id == "" || client_secret == "" || redirect_uri == "" {
		return nil, errors.New("some of required parameters are empty")
	}

	return &Client{
		client:        &http.Client{},
		clientId:      client_id,
		authorization: base64.StdEncoding.EncodeToString([]byte(client_id + ":" + client_secret)),
		redirectUri:   redirect_uri,
		scope:         scope,
	}, nil
}

func (c *Client) GenerateAuthUrl() (*string, error) {
	resource := "/authorize"
	params := url.Values{}
	params.Add("client_id", c.clientId)
	params.Add("response_type", "code")
	params.Add("redirect_uri", c.redirectUri)
	params.Add("scope", c.scope)

	u, err := url.ParseRequestURI(authUrl)
	if err != nil {
		return nil, err
	}
	u.Path = resource
	u.RawQuery = params.Encode()
	urlStr := u.String()
	return &urlStr, nil
}

func decodeTokenResponse(response *http.Response) (*OAuth2Token, error) {
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %s", response.Status)
	}

	tokenData := &oAuth2TokenResponce{}
	err := json.NewDecoder(response.Body).Decode(tokenData)
	if err != nil {
		return nil, err
	}

	expires := time.Now().Add(time.Second * time.Duration(tokenData.ExpiresIn))
	token := &OAuth2Token{
		AccessToken:  tokenData.AccessToken,
		TokenType:    tokenData.TokenType,
		Scope:        tokenData.Scope,
		Expires:      expires,
		RefreshToken: tokenData.RefreshToken,
	}

	return token, nil
}

func (c *Client) RequestAccessToken(authorization_code *string) (*OAuth2Token, error) {
	resource := "/api/token"
	u, err := url.ParseRequestURI(authUrl)
	if err != nil {
		return nil, err
	}
	u.Path = resource
	urlStr := u.String()

	request, err := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {*authorization_code},
		"redirect_uri": {c.redirectUri},
	}.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Authorization", "Basic "+c.authorization)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return decodeTokenResponse(response)
}

func (c *Client) refreshAccessToken(token *OAuth2Token) (*OAuth2Token, error) {
	resource := "/api/token"
	u, err := url.ParseRequestURI(authUrl)
	if err != nil {
		return nil, err
	}
	u.Path = resource
	urlStr := u.String()

	request, err := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {token.RefreshToken},
	}.Encode()))

	if err != nil {
		return nil, err
	}

	request.Header.Add("Authorization", "Basic "+c.authorization)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	new_token, err := decodeTokenResponse(response)
	if err != nil {
		return nil, err
	}
	if new_token.RefreshToken == "" {
		new_token.RefreshToken = token.RefreshToken
	}

	return new_token, nil
}

type Artist struct {
	Href string `json:"href"`
	Id   string `json:"id"`
	Name string `json:"name"`
}

type FollowedArtistsResponse struct {
	Artists struct {
		Next  *string  `json:"next"`
		Items []Artist `json:"items"`
	} `json:"artists"`
} 

func (c *Client) getFollowedArtists(token *OAuth2Token, request_limit uint) ([]Artist, error) {
	getRequestUrl := func(limit uint) (*string, error) {
		resource := "/v1/me/following"
		params := url.Values{}
		params.Add("type", "artist")
		params.Add("limit", strconv.FormatUint(uint64(limit), 10))

		u, err := url.ParseRequestURI(apiUrl)
		if err != nil {
			return nil, err
		}
		u.Path = resource
		u.RawQuery = params.Encode()
		requestUrl := u.String()
		return &requestUrl, nil
	}

	artists := make([]Artist, 0, request_limit)

	requestUrl, err := getRequestUrl(request_limit)
	if err != nil {
		return nil, err
	}

	for requestUrl != nil {

		fmt.Println(*requestUrl)

		if token.Expired() {
			refreshed_token, err := c.refreshAccessToken(token)
			if err != nil {
				return nil, err
			}
			token = refreshed_token
		}

		request, err := http.NewRequest(http.MethodGet, *requestUrl, nil)
		if err != nil {
			return nil, err
		}
		request.Header.Add("Authorization", "Bearer  "+token.AccessToken)

		response, err := c.client.Do(request)
		if err != nil {
			return nil, err
		}

		artists_group := FollowedArtistsResponse{}
		err = json.NewDecoder(response.Body).Decode(&artists_group)
		if err != nil {
			return nil, err
		}
		artists = append(artists, artists_group.Artists.Items...)
		requestUrl = artists_group.Artists.Next
	}

	return artists, nil
}

func (c *Client) GetFollowedArtists(token *OAuth2Token) ([]Artist, error) {
	return c.getFollowedArtists(token, 50)
}
