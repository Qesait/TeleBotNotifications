package telegram

import (
	"context"
	"fmt"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"TeleBotNotifications/internal/config"
	// "TeleBotNotifications/internal/logger"
)

var apiURL = "https://api.telegram.org"

type Bot struct {
	token       string
	commands    []command
	callbacks   []callback
	http_client *http.Client
	timeout     int
	ChatId      int
	lastUpdate  int
}

func NewBot(config *config.TelegramConfig) Bot {
	return Bot{
		token:       config.BotToken,
		http_client: &http.Client{},
		timeout:     config.Timeout,
		ChatId:      config.ChatId,
		lastUpdate:  0,
	}
}

func (b *Bot) HandleUpdates(ctx context.Context) error {
	updates, err := b.getNewUpdates(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		return fmt.Errorf("telegram update fetch failed with error: %w", err)
	}

	for _, update := range updates {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
			if update.Id > b.lastUpdate {
				b.lastUpdate = update.Id
			}
			if update.Message != nil {
				if !strings.HasPrefix(update.Message.Text, "/") {
					continue
				}
				go b.handleCommand(update.Message)
			} else if update.CallbackQuery != nil {
				go b.handleCallback(update.CallbackQuery)
			}
		}
	}
	return nil
}

type BotMessage struct {
	chatId                int
	Text                  string
	ParseMode             *string
	DisableWebPagePreview *bool
	DisableNotification   *bool
	ProtectContent        *bool
	ReplyMarkup           *string
}

func (m *BotMessage) BuildURL(token string) string {
	resource := fmt.Sprintf("/bot%s/sendMessage", token)
	params := url.Values{
		"chat_id": {strconv.Itoa(m.chatId)},
		"text":    {m.Text},
	}
	if m.ParseMode != nil {
		params.Add("parse_mode", *m.ParseMode)
	}
	if m.DisableWebPagePreview != nil {
		params.Add("disable_web_page_preview", strconv.FormatBool(*m.DisableWebPagePreview))
	}
	if m.DisableNotification != nil {
		params.Add("disable_notification", strconv.FormatBool(*m.DisableNotification))
	}
	if m.ProtectContent != nil {
		params.Add("protect_content", strconv.FormatBool(*m.ProtectContent))
	}
	if m.ReplyMarkup != nil {
		params.Add("reply_markup", *m.ReplyMarkup)
	}

	u, _ := url.ParseRequestURI(apiURL)
	u.Path = resource
	u.RawQuery = params.Encode()
	return u.String()
}

func (b *Bot) SendMessage(message BotMessage) error {
	message.chatId = b.ChatId
	url := message.BuildURL(b.token)
	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("sending request failed with err: %s", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("unexpected status code: %s", response.Status)
		}
		return fmt.Errorf("status code: %s, error: %s", response.Status, body)
	}

	return nil
}

func (b *Bot) Write(p []byte) (n int, err error) {
	err = b.SendMessage(BotMessage{Text: string(p)})
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// );
func ButtonRow(buttons ...string) *string {
	row := "{\"inline_keyboard\": [["
	for i, button := range buttons {
		if i != 0 {
			row = row + ","
		}
		row = row + button
	}
	row = row + "]]}"
	return &row
}

func CallbackButton(text, data string) string {
	return fmt.Sprintf("{\"text\": \"%s\",\"callback_data\": \"%s\"}", text, data)
}
func URLButton(text, url string) string {
	return fmt.Sprintf("{\"text\": \"%s\",\"url\": \"%s\"}", text, url)
}
