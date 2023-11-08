package telegram

import (
	"TeleBotNotifications/internal/config"
	"TeleBotNotifications/internal/logger"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var apiURL = "https://api.telegram.org"

type Bot struct {
	token       string
	commands    []command
	http_client *http.Client
	updateDelay time.Duration
	adminChatId int
	lastUpdate  int
}

func NewBot(config *config.TelegramConfig) Bot {
	return Bot{
		token:       config.BotToken,
		http_client: &http.Client{},
		updateDelay: time.Duration(config.UpdateDelay) * time.Second,
		adminChatId: config.AdminChatId,
		lastUpdate:  0,
	}
}

type ReceivedMessage struct {
	updateId int
	UserId   int
	ChatId   int
	Text     string
}

func (b *Bot) Run(port uint) {
	err := b.UpdateCommands()
	if err != nil {
		logger.Error.Println(err)
	}

	logger.General.Println("Telegram bot started")

	for {
		time.Sleep(b.updateDelay)
		messages, err := b.getNewMessages()
		if err != nil {
			logger.Error.Println("Error getting new messages: ", err)
			continue
		}
		if len(messages) == 0 {
			continue
		}

		for i := 0; i < len(messages); i++ {
			if messages[i].updateId > b.lastUpdate {
				b.lastUpdate = messages[i].updateId
			}
			if !strings.HasPrefix(messages[i].Text, "/") {
				continue
			}
			b.handleCommand(messages[i])
		}

	}
}

func (b *Bot) handleCommand(message ReceivedMessage) {
	for j := 0; j < len(b.commands); j++ {
		if strings.HasPrefix(message.Text, b.commands[j].Keyword) {
			logger.General.Println("<-" + message.Text)
			message.Text = strings.TrimSpace(strings.TrimPrefix(message.Text, b.commands[j].Keyword))
			b.commands[j].Handler(message)
			return
		}
	}
}

type BotMessage struct {
	ChatId                int
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
		"chat_id": {strconv.Itoa(m.ChatId)},
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
	response, err := http.Get(message.BuildURL(b.token))
	if err != nil {
		// logger.Error.Println("telegram bot failed to send message with error", err)
		return fmt.Errorf("sending request failed with err: %s", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		// logger.Error.Println("telegram bot failed to send message because of unexpected status code", response)
		return fmt.Errorf("unexpected status code: %s", response.Status)
	}

	return nil
}

func (b *Bot) Write(p []byte) (n int, err error) {
	err = b.SendMessage(BotMessage{ChatId: b.adminChatId, Text: string(p)})
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
