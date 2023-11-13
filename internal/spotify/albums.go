package spotify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type getAlbumTracksResponse struct {
	Href     string            `json:"href"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
	Next     *string           `json:"next"`
	Previous *string           `json:"previous"`
	Total    int               `json:"total"`
	Tracks   []SimplifiedTrack `json:"items"`
}

// TODO: arguments does not make sense
func (c *Client) GetAlbumTracks(token *OAuth2Token, albumId string, limit, offset uint64, market *string) ([]SimplifiedTrack, error) {
	if limit < 1 || 50 > limit {
		return nil, fmt.Errorf("limit %d is out range 1-50", limit)
	}
	params := url.Values{
		"id":     {albumId},
		"limit":  {strconv.FormatUint(limit, 10)},
		"offset": {strconv.FormatUint(offset, 10)},
	}
	if market != nil {
		params.Add("market", *market)
	}

	queueResource := fmt.Sprintf("/v1/albums/%s/tracks", albumId)
	tmp := fmt.Sprintf("%s%s?%s", apiUrl, queueResource, params.Encode())
	requestURL := &tmp

	tracks := make([]SimplifiedTrack, 0, limit)

	for requestURL != nil {
		if err := c.checkToken(token); err != nil {
			return nil, err
		}

		request, err := http.NewRequest(http.MethodGet, *requestURL, nil)
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

		var tracksPart getAlbumTracksResponse
		err = json.NewDecoder(response.Body).Decode(&tracksPart)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, tracksPart.Tracks...)
		requestURL = tracksPart.Next
	}
	return tracks, nil
}
