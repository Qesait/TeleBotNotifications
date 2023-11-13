package spotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) AddItemtoPlaybackQueue(token *OAuth2Token, uri, deviceId *string) error {
	const queueResource = "/v1/me/player/queue"
	params := url.Values{"uri": {*uri}}
	if deviceId != nil {
		params.Add("device_id", *deviceId)
	}
	requestURL := fmt.Sprintf("%s%s?%s", apiUrl, queueResource, params.Encode())

	request, err := http.NewRequest(http.MethodPost, requestURL, nil)
	if err != nil {
		return err
	}

	if err := c.checkToken(token); err != nil {
		return err
	}
	request.Header.Add("Authorization", "Bearer  "+token.AccessToken)
	response, err := c.client.Do(request)
	if err != nil || response.StatusCode != http.StatusNoContent {
		explanation := &errorResponse{}
		if err := json.NewDecoder(response.Body).Decode(explanation); err != nil {
			return fmt.Errorf("http error %s, cant  decode response %s", response.Status, err)
		}
		return fmt.Errorf("http request fail: %s, %s", response.Status, explanation.Error.Message)
	}

	return nil
}

// uris, ofset and position_ms not implemented
func (c *Client) StartResumePlayback(token *OAuth2Token, contextURI, deviceId *string) error {
	const queueResource = "/v1/me/player/play"
	params := url.Values{}
	requestURL := fmt.Sprintf("%s%s", apiUrl, queueResource)
	if deviceId != nil {
		params.Add("device_id", *deviceId)
		requestURL = requestURL + "?" + params.Encode()
	}

	var request *http.Request
	var err error
	if contextURI == nil {
		request, err = http.NewRequest(http.MethodPut, requestURL, nil)
	} else {
		requestBody := map[string]interface{}{"context_uri": *contextURI}
		var jsonData []byte
		jsonData, err = json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("error encoding JSON: %s", err)
		}
		request, err = http.NewRequest(http.MethodPut, requestURL, bytes.NewBuffer(jsonData))
	}
	if err != nil {
		return err
	}

	if err := c.checkToken(token); err != nil {
		return err
	}
	request.Header.Add("Authorization", "Bearer  "+token.AccessToken)
	request.Header.Add("Content-Type", "application/json")

	response, err := c.client.Do(request)
	if err != nil || response.StatusCode != http.StatusNoContent {
		explanation := &errorResponse{}
		if err := json.NewDecoder(response.Body).Decode(explanation); err != nil {
			return fmt.Errorf("http error %s, cant  decode response %s", response.Status, err)
		}
		return fmt.Errorf("http request fail: %s, %s", response.Status, explanation.Error.Message)
	}

	return nil
}

