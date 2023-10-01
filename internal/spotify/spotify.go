package spotify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	// "io"

	"TeleBotNotifications/internal/config"
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

func NewClient(config config.SpotifyConfig) (*Client, error) {
	if config.ClientId == "" || config.ClientSecret == "" {
		return nil, fmt.Errorf("credentials required")
	}

	return &Client{
		client:        &http.Client{},
		clientId:      config.ClientId,
		authorization: base64.StdEncoding.EncodeToString([]byte(config.ClientId + ":" + config.ClientSecret)),
		redirectUri:   config.RedirectUri,
		scope:         config.Scope,
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

type Album struct {
	Id          string
	Name        string
	AlbumType   string
	AlbumGroup  string
	Url         string
	ImageUrl    string
	ReleaseDate time.Time
	Artists     []Artist
}

type image struct {
	Url    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type album struct {
	Id           string `json:"id"`
	AlbumType    string `json:"album_type"`
	ExternalUrls struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Images               []image  `json:"images"`
	Name                 string   `json:"name"`
	ReleaseDate          string   `json:"release_date"`
	ReleaseDatePrecision string   `json:"release_date_precision"`
	Artists              []Artist `json:"artists"`
	AlbumGroup           string   `json:"album_group"`
}

type getArtistAlbumsResponse struct {
	Next   *string  `json:"next"`
	Albums []album `json:"items"`
}

func decodeAlbulmsResponse(response *http.Response) ([]Album, *string, error) {
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("unexpected status code: %s", response.Status)
	}

	responseData := &getArtistAlbumsResponse{}
	err := json.NewDecoder(response.Body).Decode(responseData)
	if err != nil {
		return nil, nil, err
	}

	albums := make([]Album, 0, len(responseData.Albums))
	for i := 0; i < len(responseData.Albums); i++ {
		var t time.Time
		if responseData.Albums[i].ReleaseDatePrecision == "day" {
			t, err = time.Parse("2006-01-02", responseData.Albums[i].ReleaseDate)
			if err != nil {
				return nil, nil, fmt.Errorf("error parsing date: %s", err)
			}
		} else if responseData.Albums[i].ReleaseDatePrecision == "month" {
			t, err = time.Parse("2006-01", responseData.Albums[i].ReleaseDate)
			if err != nil {
				return nil, nil, fmt.Errorf("error parsing date: %s", err)
			}
		} else if responseData.Albums[i].ReleaseDatePrecision == "year" {
			t, err = time.Parse("2006", responseData.Albums[i].ReleaseDate)
			if err != nil {
				return nil, nil, fmt.Errorf("error parsing date: %s", err)
			}
		}

		album := Album{
			Id:          responseData.Albums[i].Id,
			Name:        responseData.Albums[i].Name,
			AlbumType:   responseData.Albums[i].AlbumType,
			AlbumGroup:  responseData.Albums[i].AlbumGroup,
			Url:         responseData.Albums[i].ExternalUrls.Spotify,
			// ImageUrl:    responseData.Albums[i].Images[0].Url,
			ReleaseDate: t,
			Artists:     responseData.Albums[i].Artists,
		}
		if len(responseData.Albums[i].Images) > 0 {
			album.ImageUrl = responseData.Albums[i].Images[0].Url
		}
		albums = append(albums, album)
	}

	return albums, responseData.Next, nil
}

func (c *Client) getArtistAlbums(token *OAuth2Token, artistId string, include_groups string, requestLimit uint) ([]Album, error) {
	getRequestUrl := func() (*string, error) {
		params := url.Values{
			"include_groups": {include_groups},
			"limit":          {strconv.FormatUint(uint64(requestLimit), 10)},
		}

		u, err := url.ParseRequestURI(apiUrl)
		if err != nil {
			return nil, err
		}
		u.Path = fmt.Sprintf("/v1/artists/%s/albums", artistId)
		u.RawQuery = params.Encode()
		requestUrl := u.String()
		return &requestUrl, nil
	}

	albums := make([]Album, 0, requestLimit)

	requestUrl, err := getRequestUrl()
	if err != nil {
		return nil, err
	}

	for requestUrl != nil {
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

		var albumsPart []Album
		albumsPart, requestUrl, err = decodeAlbulmsResponse(response)
		if err != nil {
			return nil, err
		}

		albums = append(albums, albumsPart...)
	}
	return albums, nil
}

//album,single,compilation,appears_on
func (c *Client) GetArtistAlbums(token *OAuth2Token, artist *Artist) ([]Album, error) {
	return c.getArtistAlbums(token, artist.Id, "album,single", 50)
}