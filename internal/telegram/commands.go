package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type command struct {
	Keyword     string         `json:"command"`
	Description string         `json:"description"`
	Handler     CommandHandler `json:"-"`
}

// TODO: type CommandHandler func(ReceivedMessage) error
type CommandHandler func(ReceivedMessage)

func (b *Bot) AddCommand(keyword string, description string, handler CommandHandler) {
	b.commands = append(b.commands, command{
		Keyword:     "/" + keyword,
		Description: description,
		Handler:     handler,
	})
}

func (b *Bot) UpdateCommands() error {
	resource := fmt.Sprintf("/bot%s/setMyCommands", b.token)
	u, err := url.ParseRequestURI(apiURL)
	if err != nil {
		return err
	}
	u.Path = resource
	requestUrl := u.String()

	// Define the payload data as a Go struct.
	data := map[string][]command{
		"commands": b.commands,
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %i", err)
	}

	// Create a request with the payload.
	request, err := http.NewRequest(http.MethodPost, requestUrl, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("error creating request: %i", err)
	}
	request.Header.Set("Content-Type", "application/json")

	// Send the request.
	response, err := b.http_client.Do(request)
	if err != nil {
		return fmt.Errorf("error sending request: %i", err)
	}
	defer response.Body.Close()

	// Check the response.
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %s", response.Status)
	}

	return nil
}

type ReceivedMessage struct {
	UserId int
	ChatId int
	Text   string
}

func (b *Bot) handleCommand(m *message) {
	for j := 0; j < len(b.commands); j++ {
		if strings.HasPrefix(m.Text, b.commands[j].Keyword) {
			b.commands[j].Handler(ReceivedMessage{
				UserId: m.From.Id,
				ChatId: m.Chat.Id,
				Text:   strings.TrimSpace(strings.TrimPrefix(m.Text, b.commands[j].Keyword)),
			})
			return
		}
	}
}
