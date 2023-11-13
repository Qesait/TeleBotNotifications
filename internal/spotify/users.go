package spotify

import (
	"encoding/json"
	"net/http"
	"net/url"
	"fmt"
	"strconv"
)

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
		if err != nil || response.StatusCode != http.StatusOK {
			explanation := &errorResponse{}
			if err := json.NewDecoder(response.Body).Decode(explanation); err != nil {
				return nil, fmt.Errorf("http error %s, cant  decode response %s", response.Status, err)
			}
			return nil, fmt.Errorf("http request fail: %s, %s", response.Status, explanation.Error.Message)
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