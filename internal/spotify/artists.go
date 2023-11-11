package spotify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)




func (c *Client) GetFollowedArtists(token *OAuth2Token) ([]Artist, error) {
	return c.getFollowedArtists(token, 50)
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
	Uri                  string   `json:"uri"`
	Artists              []Artist `json:"artists"`
	AlbumGroup           string   `json:"album_group"`
}

type getArtistAlbumsResponse struct {
	Next   *string `json:"next"`
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
			Id:         responseData.Albums[i].Id,
			Name:       responseData.Albums[i].Name,
			AlbumType:  responseData.Albums[i].AlbumType,
			AlbumGroup: responseData.Albums[i].AlbumGroup,
			Url:        responseData.Albums[i].ExternalUrls.Spotify,
			Uri:        responseData.Albums[i].Uri,
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

// album,single,compilation,appears_on
func (c *Client) GetArtistAlbums(token *OAuth2Token, artist *Artist) ([]Album, error) {
	return c.getArtistAlbums(token, artist.Id, "album,single", 50)
}