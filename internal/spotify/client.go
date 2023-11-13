package spotify

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"TeleBotNotifications/internal/config"
)


type Client struct {
	client        *http.Client
	clientId      string
	authorization string
	redirectUri   string
	scope         string
}

func NewClient(config *config.SpotifyConfig) (*Client, error) {
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