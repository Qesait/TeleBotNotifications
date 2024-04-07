package telegram

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"context"
	"errors"
	"strconv"
)

type user struct {
	Id           int    `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

type message struct {
	MessageId int  `json:"message_id"`
	From      user `json:"from"`
	Chat      struct {
		Id        int    `json:"id"`
		FirstName string `json:"first_name"`
		Username  string `json:"username"`
		Type      string `json:"type"`
	} `json:"chat"`
	Date int    `json:"date"`
	Text string `json:"text"`
}

type callbackQuery struct {
	Id   string `json:"id"`
	From user   `json:"from"`
	// message
	InlineMessageId *string `json:"inline_message_id"`
	ChatInstance    string `json:"chat_instance"`
	Data            *string `json:"data"`
	GameShortName   *string `json:"game_short_name"`
}

type update struct {
	Id            int            `json:"update_id"`
	Message       *message       `json:"message"`
	CallbackQuery *callbackQuery `json:"callback_query"`
}

type updateResponse struct {
	Ok     bool     `json:"ok"`
	Result []update `json:"result"`
}

func (b *Bot) createUpdateUrl() string {
	resource := fmt.Sprintf("/bot%s/getUpdates", b.token)
	params := url.Values{
		"offset":          {strconv.Itoa(b.lastUpdate + 1)},
		"timeout":         {strconv.Itoa(b.timeout)},
		"allowed_updates": {"[\"message\", \"callback_query\"]"},
	}

	u, _ := url.ParseRequestURI(apiURL)
	u.Path = resource
	u.RawQuery = params.Encode()
	return u.String()
}

// https://core.telegram.org/bots/api#getupdates
func (b *Bot) getNewUpdates(ctx context.Context) ([]update, error) {
	requestUrl := b.createUpdateUrl()
    req, err := http.NewRequestWithContext(ctx, "GET", requestUrl, nil)
    if err != nil {
        return nil, fmt.Errorf("creating request failed with err: %s", err)
    }

    response, err := b.http_client.Do(req)
    if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("request was canceled: %w", err)
		}
        return nil, fmt.Errorf("sending request failed with err: %s", err)
    }
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %s", response.Status)
	}

	updates := &updateResponse{}
	err = json.NewDecoder(response.Body).Decode(updates)
	if err != nil {
		return nil, fmt.Errorf("decoding as message failed with error: %s", err)
	}

	return updates.Result, nil
}
