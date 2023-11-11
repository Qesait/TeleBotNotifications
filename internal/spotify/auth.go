package spotify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

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
