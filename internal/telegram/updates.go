package telegram

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type update struct {
	UpdateId int `json:"update_id"`
	Message  struct {
		MessageId int `json:"message_id"`
		From      struct {
			Id           int    `json:"id"`
			IsBot        bool   `json:"is_bot"`
			FirstName    string `json:"first_name"`
			Username     string `json:"username"`
			LanguageCode string `json:"language_code"`
		} `json:"from"`
		Chat struct {
			Id        int    `json:"id"`
			FirstName string `json:"first_name"`
			Username  string `json:"username"`
			Type      string `json:"type"`
		} `json:"chat"`
		Date int    `json:"date"`
		Text string `json:"text"`
	} `json:"message"`
}

type updateResponse struct {
	Ok     bool     `json:"ok"`
	Result []update `json:"result"`
}

func extract(update_with_message update) ReceivedMessage {
	return ReceivedMessage{
		updateId: update_with_message.UpdateId,
		UserId:   update_with_message.Message.From.Id,
		ChatId:   update_with_message.Message.Chat.Id,
		Text:     update_with_message.Message.Text,
	}
}

func (b *Bot) createUpdateUrl() string {
	resource := fmt.Sprintf("/bot%s/getUpdates", b.token)
	params := url.Values{
		"offset":          {strconv.Itoa(b.lastUpdate + 1)},
		"allowed_updates": {"[\"message\"]"},
	}

	u, _ := url.ParseRequestURI(apiURL)
	u.Path = resource
	u.RawQuery = params.Encode()
	return u.String()
}

func (b *Bot) getNewMessages() ([]ReceivedMessage, error) {
	requestUrl := b.createUpdateUrl()
	response, err := b.http_client.Get(requestUrl)
	if err != nil {
		return nil, fmt.Errorf("sending request failed with err: %s", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %s", response.Status)
	}

	updates := &updateResponse{}
	err = json.NewDecoder(response.Body).Decode(updates)
	if err != nil {
		return nil, fmt.Errorf("decoding failed with error: %s", err)
	}

	messages := make([]ReceivedMessage, 0, len(updates.Result))
	for i := 0; i < len(updates.Result); i++ {
		messages = append(messages, extract(updates.Result[i]))
	}

	return messages, nil
}
