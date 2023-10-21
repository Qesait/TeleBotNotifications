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

type Message struct {
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

func (b *Bot) handleCommand(message Message) {
	for j := 0; j < len(b.commands); j++ {
		if strings.HasPrefix(message.Text, b.commands[j].Keyword) {
			logger.General.Println("<-" + message.Text)
			message.Text = strings.TrimSpace(strings.TrimPrefix(message.Text, b.commands[j].Keyword))
			b.commands[j].Handler(message)
			return
		}
	}
}

func (b *Bot) SendMessage(message string, to int) error {
	resource := fmt.Sprintf("/bot%s/sendMessage", b.token)
	params := url.Values{
		"chat_id": {strconv.Itoa(to)},
		"text":    {message},
	}

	u, _ := url.ParseRequestURI(apiURL)
	u.Path = resource
	u.RawQuery = params.Encode()
	requestUrl := u.String()

	response, err := http.Get(requestUrl)
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
	err = b.SendMessage(string(p), b.adminChatId)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
