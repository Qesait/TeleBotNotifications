package telegram

import (
	"TeleBotNotifications/internal/config"
	"TeleBotNotifications/internal/logger"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var apiURL = "https://api.telegram.org"

type Bot struct {
	token       string
	commands    []command
	http_client *http.Client
	timeout     int
	adminChatId int
	lastUpdate  int
}

func NewBot(config *config.TelegramConfig) Bot {
	return Bot{
		token:       config.BotToken,
		http_client: &http.Client{},
		timeout:     config.Timeout,
		adminChatId: config.AdminChatId,
		lastUpdate:  0,
	}
}

type ReceivedMessage struct {
	UserId int
	ChatId int
	Text   string
}

func (b *Bot) Run(port uint) {
	err := b.UpdateCommands()
	if err != nil {
		logger.Error.Println(err)
	}

	logger.General.Println("Telegram bot started")

	for {
		updates, err := b.getNewUpdates()
		if err != nil {
			logger.Error.Println("telegram update fetch failed with error: ", err)
			continue
		}

		for _, update := range updates {
			if update.Id > b.lastUpdate {
				b.lastUpdate = update.Id
			}
			if update.Message != nil {
				if !strings.HasPrefix(update.Message.Text, "/") {
					continue
				}
				b.handleCommand(update.Message)
			} else if update.CallbackQuery != nil {
				err := b.answerCallbackQuery(update.CallbackQuery.Id)
				if err != nil {
					logger.Error.Println("telegram bot failed to answer callback query: ", err)
				}
			}
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
		return fmt.Errorf("sending request failed with err: %s", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
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

func (b *Bot) answerCallbackQuery(queryId string) error {
	resourse := fmt.Sprintf("/bot%s/answerCallbackQuery", b.token)
	params := url.Values{
		"callback_query_id": {queryId},
	}
	u, _ := url.ParseRequestURI(apiURL)
	u.Path = resourse
	u.RawQuery = params.Encode()
	requestURL := u.String()
	_, err := http.Get(requestURL)
	if err != nil {
		return fmt.Errorf("sending request failed with err: %s", err)
	}
	return nil
}
