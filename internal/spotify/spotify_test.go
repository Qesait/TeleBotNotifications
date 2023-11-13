package spotify

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"TeleBotNotifications/internal/config"
)

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (s roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return s(r)
}

// Check request parameters?
func newClient(t *testing.T, method string, statusCode int, path string, body string) *Client {
	return &Client{
		client: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if path != r.URL.Path {
					t.Error("Expected request path", path, "got", r.URL.Path)
				}
				if method != r.Method {
					t.Error("Expected request method", method, "got", r.Method)
				}

				return &http.Response{
					StatusCode: statusCode,
					Body:       io.NopCloser(strings.NewReader(body)),
				}, nil
			}),
		},
		clientId:      "id",
		authorization: "id:secret",
		redirectUri:   "uri",
		scope:         "scope",
	}
}

func Test_GenerateAuthUrl(t *testing.T) {
	type args struct {
		client_id    string
		redirect_uri string
		scope        string
	}

	want := func(args args) string {
		return fmt.Sprintf("https://accounts.spotify.com/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s", args.client_id, args.redirect_uri, args.scope)
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Ok",
			args: args{
				client_id:    "123",
				redirect_uri: "adress",
				scope:        "everything",
			},
			wantErr: false,
		},
		{
			name: "Empty id",
			args: args{
				client_id:    "",
				redirect_uri: "adress",
				scope:        "everything",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(&config.SpotifyConfig{
				ClientId:     tt.args.client_id,
				ClientSecret: "secret",
				Scope:        tt.args.scope,
				RedirectUri:  tt.args.redirect_uri})
			if err != nil {
				if !tt.wantErr {
					t.Error(err)
				}
			} else {
				got, err := client.GenerateAuthUrl()
				if tt.wantErr && err == nil {
					t.Error("error expected")
				} else {
					if err != nil {
						t.Error(err)
					} else if expected := want(tt.args); expected != *got {
						t.Error("\nexpected\t", expected, "\ngot\t\t", *got)
					}
				}
			}
		})
	}
}

func Test_decodeTokenResponse(t *testing.T) {
	type args struct {
		response       string
		statusCode     int
		expected_token OAuth2Token
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Ok",
			args: args{
				response: `
				{
					"access_token": "sample-access-token",
					"token_type": "bearer",
					"scope": "read write",
					"expires_in": 3600,
					"refresh_token": "sample-refresh-token"
				}
				`,
				statusCode: http.StatusOK,
				expected_token: OAuth2Token{
					AccessToken:  "sample-access-token",
					TokenType:    "bearer",
					Scope:        "read write",
					Expires:      time.Now().Add(time.Hour),
					RefreshToken: "sample-refresh-token",
				},
			},
			wantErr: false,
		},
		{
			name: "StatusCodeNotOK",
			args: args{
				response: `
				{
					"access_token": "sample-access-token",
					"token_type": "bearer",
					"scope": "read write",
					"expires_in": 3600,
					"refresh_token": "sample-refresh-token"
				}
				`,
				statusCode:     http.StatusBadRequest,
				expected_token: OAuth2Token{},
			},
			wantErr: true,
		},
		{
			name: "Badresponse",
			args: args{
				response:       "definetely bad response",
				statusCode:     http.StatusOK,
				expected_token: OAuth2Token{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResponse := httptest.NewRecorder()
			mockResponse.WriteHeader(tt.args.statusCode)
			mockResponse.WriteString(tt.args.response)
			response := mockResponse.Result()

			token, err := decodeTokenResponse(response)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("Error decoding token response: %v", err)
			} else if tt.wantErr {
				t.Error("Error expected")
			}

			if token.AccessToken != tt.args.expected_token.AccessToken ||
				token.TokenType != tt.args.expected_token.TokenType ||
				token.Scope != tt.args.expected_token.Scope ||
				token.Expires.Sub(tt.args.expected_token.Expires).Abs() > time.Second ||
				token.RefreshToken != tt.args.expected_token.RefreshToken {
				t.Errorf("Tokens do not match. \nExpected: \t%+v, \nGot: \t%+v", &tt.args.expected_token, token)
			}
		})
	}
}

func Test_RequestAccessToken(t *testing.T) {
	type args struct {
		authorization_code string
		statusCode         int
		response           string
		expected_request   string
		expected_token     OAuth2Token
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Ok",
			args: args{
				authorization_code: "authorization-code",
				statusCode:         http.StatusOK,
				response: `
				{
					"access_token": "sample-access-token",
					"token_type": "bearer",
					"scope": "read write",
					"expires_in": 3600,
					"refresh_token": "sample-refresh-token"
				}
				`,
				expected_request: "/api/token",
				expected_token: OAuth2Token{
					AccessToken:  "sample-access-token",
					TokenType:    "bearer",
					Scope:        "read write",
					Expires:      time.Now().Add(time.Hour),
					RefreshToken: "sample-refresh-token",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			client := newClient(t, http.MethodPost, tt.args.statusCode, tt.args.expected_request, tt.args.response)

			token, err := client.RequestAccessToken(&tt.args.authorization_code)

			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("Error decoding token response: %v", err)
			} else if tt.wantErr {
				t.Error("Error expected")
			}

			if token.AccessToken != tt.args.expected_token.AccessToken ||
				token.TokenType != tt.args.expected_token.TokenType ||
				token.Scope != tt.args.expected_token.Scope ||
				token.Expires.Sub(tt.args.expected_token.Expires).Abs() > time.Second ||
				token.RefreshToken != tt.args.expected_token.RefreshToken {
				t.Errorf("Tokens do not match. \nExpected: \t%+v, \nGot: \t%+v", &tt.args.expected_token, token)
			}
		})
	}
}

func Test_refreshAccessToken(t *testing.T) {
	type args struct {
		current_token    OAuth2Token
		statusCode       int
		response         string
		expected_request string
		expected_token   OAuth2Token
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "NoNewRefreshToken",
			args: args{
				current_token: OAuth2Token{
					AccessToken:  "sample-access-token",
					TokenType:    "bearer",
					Scope:        "read write",
					Expires:      time.Now().Add(-time.Hour),
					RefreshToken: "sample-refresh-token",
				},
				statusCode: http.StatusOK,
				response: `
				{
					"access_token": "sample-access-token",
					"token_type": "bearer",
					"scope": "read write",
					"expires_in": 3600
				}
				`,
				expected_request: "/api/token",
				expected_token: OAuth2Token{
					AccessToken:  "sample-access-token",
					TokenType:    "bearer",
					Scope:        "read write",
					Expires:      time.Now().Add(time.Hour),
					RefreshToken: "sample-refresh-token",
				},
			},
			wantErr: false,
		},
		{
			name: "NewRefreshToken",
			args: args{
				current_token: OAuth2Token{
					AccessToken:  "sample-access-token",
					TokenType:    "bearer",
					Scope:        "read write",
					Expires:      time.Now().Add(-time.Hour),
					RefreshToken: "sample-refresh-token",
				},
				statusCode: http.StatusOK,
				response: `
				{
					"access_token": "sample-access-token",
					"token_type": "bearer",
					"scope": "read write",
					"expires_in": 3600,
					"refresh_token": "new-sample-refresh-token"
				}
				`,
				expected_request: "/api/token",
				expected_token: OAuth2Token{
					AccessToken:  "sample-access-token",
					TokenType:    "bearer",
					Scope:        "read write",
					Expires:      time.Now().Add(time.Hour),
					RefreshToken: "new-sample-refresh-token",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			client := newClient(t, http.MethodPost, tt.args.statusCode, tt.args.expected_request, tt.args.response)

			token, err := client.refreshAccessToken(&tt.args.current_token)

			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("Error decoding token response: %v", err)
			} else if tt.wantErr {
				t.Error("Error expected")
			}

			if token.AccessToken != tt.args.expected_token.AccessToken ||
				token.TokenType != tt.args.expected_token.TokenType ||
				token.Scope != tt.args.expected_token.Scope ||
				token.Expires.Sub(tt.args.expected_token.Expires).Abs() > time.Second ||
				token.RefreshToken != tt.args.expected_token.RefreshToken {
				t.Errorf("Tokens do not match. \nExpected: \t%+v, \nGot: \t%+v", &tt.args.expected_token, token)
			}
		})
	}
}

func Test_getFollowedArtists(t *testing.T) {
	type args struct {
		current_token    OAuth2Token
		statusCode       int
		response         string
		expected_request string
		expected_artists []Artist
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				current_token: OAuth2Token{
					AccessToken:  "sample-access-token",
					TokenType:    "bearer",
					Scope:        "read write",
					Expires:      time.Now().Add(time.Hour),
					RefreshToken: "sample-refresh-token",
				},
				statusCode: http.StatusOK,
				response: `
				{
					"artists":{
					  "href":"string",
					  "limit":0,
					  "next": null,
					  "cursors":{
						"after":"string",
						"before":"string"
					  },
					  "total":0,
					  "items":[
						{
						  "external_urls":{
							"spotify":"string"
						  },
						  "followers":{
							"href":"string",
							"total":0
						  },
						  "genres":[
							"Prog rock",
							"Grunge"
						  ],
						  "href":"href-1",
						  "id":"1",
						  "images":[
							{
							  "url":"some-uri",
							  "height":300,
							  "width":300
							}
						  ],
						  "name":"artist-1",
						  "popularity":0,
						  "type":"artist",
						  "uri":"string"
						},
						{
						  "external_urls":{
							"spotify":"string"
						  },
						  "followers":{
							"href":"string",
							"total":0
						  },
						  "genres":[
							"Prog rock",
							"Grunge"
						  ],
						  "href":"href-2",
						  "id":"2",
						  "images":[
							{
							  "url":"some-uri",
							  "height":300,
							  "width":300
							}
						  ],
						  "name":"artist-2",
						  "popularity":0,
						  "type":"artist",
						  "uri":"string"
						}
					  ]
					}
				}
				`,
				expected_request: "/v1/me/following",
				expected_artists: []Artist{
					{"1", "artist-1"},
					{"2", "artist-2"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			client := newClient(t, http.MethodGet, tt.args.statusCode, tt.args.expected_request, tt.args.response)

			artists, err := client.getFollowedArtists(&tt.args.current_token, 2)

			if err != nil {
				if !tt.wantErr {
					t.Errorf("Error decoding response: %v", err)
				}
				return
			} else if tt.wantErr {
				t.Error("Error expected")
				return
			}

			if len(artists) != len(tt.args.expected_artists) {
				t.Errorf("Wrond result. \nExpected: \t%+v, \nGot: \t%+v", &tt.args.expected_artists, artists)
			}

			for i := 0; i < len(tt.args.expected_artists); i++ {
				if tt.args.expected_artists[i] != artists[i] {
					t.Errorf("Wrond result. \nExpected: \t%+v, \nGot: \t%+v", &tt.args.expected_artists, artists)
				}
			}
		})
	}
}

func Test_getArtistAlbums(t *testing.T) {
	type args struct {
		current_token    OAuth2Token
		statusCode       int
		response         string
		expected_request string
		expected_result  []Album
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				current_token: OAuth2Token{
					AccessToken:  "sample-access-token",
					TokenType:    "bearer",
					Scope:        "read write",
					Expires:      time.Now().Add(time.Hour),
					RefreshToken: "sample-refresh-token",
				},
				statusCode: http.StatusOK,
				response: `
				{
					"limit": 20,
					"next": null,
					"total": 4,
					"items": [
					  {
						"album_type": "album",
						"total_tracks": 9,
						"external_urls": {
						  "spotify": "spotify_url"
						},
						"id": "2up3OPMp9Tb4dAKM2er111",
						"images": [
						  {
							"url": "image-url-1",
							"height": 300,
							"width": 300
						  }
						],
						"name": "name-1",
						"release_date": "2001",
						"release_date_precision": "year",
						"type": "album",
						"uri": "spotify:album:1up",
						"artists": [
						  {
							"external_urls": {
							  "spotify": "string"
							},
							"href": "string",
							"id": "1",
							"name": "name-1",
							"type": "artist",
							"uri": "string"
						  }
						],
						"album_group": "compilation"
					  },
					  {
						"album_type": "compilation",
						"total_tracks": 9,
						"external_urls": {
						  "spotify": "another_spotify_url"
						},
						"id": "2up3OPMp9Tb4dAKM2erWXQ",
						"images": [
						  {
							"url": "image-url-2",
							"height": 300,
							"width": 300
						  }
						],
						"name": "name-2",
						"release_date": "1981-12-01",
						"release_date_precision": "day",
						"type": "album",
						"uri": "spotify:album:2up",
						"artists": [
						  {
							"external_urls": {
							  "spotify": "string"
							},
							"href": "string",
							"id": "2",
							"name": "name-2",
							"type": "artist",
							"uri": "string"
						  }
						],
						"album_group": "compilation"
					  }
					]
				  }
				`,
				expected_request: "/v1/artists/id/albums",
				expected_result: []Album{
					{"2up3OPMp9Tb4dAKM2er111", "name-1", "album", "compilation", "spotify_url", "spotify:album:1up", "image-url-1", time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC), []Artist{{"1", "name-1"}}},
					{"2up3OPMp9Tb4dAKM2erWXQ", "name-2", "compilation", "compilation", "another_spotify_url", "spotify:album:2up", "image-url-2", time.Date(1981, 12, 1, 0, 0, 0, 0, time.UTC), []Artist{{"2", "name-2"}}},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			client := newClient(t, http.MethodGet, tt.args.statusCode, tt.args.expected_request, tt.args.response)

			albums, err := client.getArtistAlbums(&tt.args.current_token, "id", "", 2)

			if err != nil {
				if !tt.wantErr {
					t.Errorf("Error decoding response: %v", err)
				}
				return
			} else if tt.wantErr {
				t.Error("Error expected")
				return
			}

			if len(albums) != len(tt.args.expected_result) {
				t.Errorf("Wrond result. \nExpected: \t%+v, \nGot: \t%+v", &tt.args.expected_result, albums)
				return
			}

			for i := 0; i < len(tt.args.expected_result); i++ {
				if !compareAlbums(tt.args.expected_result[i], albums[i]) {
					t.Errorf("Wrond result. \nExpected: \t%+v, \nGot: \t%+v", &tt.args.expected_result, albums)
					break
				}
			}
		})
	}
}

func compareAlbums(a1, a2 Album) bool {
	if a1.Id != a2.Id ||
		a1.Name != a2.Name ||
		a1.AlbumGroup != a2.AlbumGroup ||
		a1.AlbumType != a2.AlbumType ||
		a1.Url != a2.Url ||
		a1.ImageUrl != a2.ImageUrl ||
		a1.ReleaseDate != a2.ReleaseDate {
		return false
	}
	if len(a1.Artists) != len(a2.Artists) {
		return false
	}
	for i, a := range a1.Artists {
		if a != a2.Artists[i] {
			return false
		}
	}
	return true
}
