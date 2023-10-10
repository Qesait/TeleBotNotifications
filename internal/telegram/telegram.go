package telegram

import (
	"TeleBotNotifications/internal/config"
	"TeleBotNotifications/internal/logger"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type CommandHandler func(Message)

type command struct {
	Keyword string
	Handler CommandHandler
}

type Bot struct {
	token       string
	commands    []command
	http_client *http.Client
	updateDelay time.Duration
	adminChatId int
}

func NewBot(config *config.TelegramConfig) Bot {
	return Bot{
		token:       config.BotToken,
		http_client: &http.Client{},
		updateDelay: time.Duration(config.UpdateDelay) * time.Second,
		adminChatId: config.AdminChatId,
	}
}

// TODO: Add command description to the bot
func (b *Bot) AddCommand(keyword string, handler CommandHandler) {
	b.commands = append(b.commands, command{"/" + keyword, handler})
}

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
		// Entities struct {
		// 	Type string `json:"type"`
		// 	// дальше мне лень
		// } `json:"entities"`
	} `json:"message"`
}

type updateResponse struct {
	Ok     bool     `json:"ok"`
	Result []update `json:"result"`
}

type message struct {
	UpdateId int
	From     int
	Chat     int
	Text     string
}

func extract(update_with_message update) message {
	return message{
		UpdateId: update_with_message.UpdateId,
		From:     update_with_message.Message.From.Id,
		Chat:     update_with_message.Message.Chat.Id,
		Text:     update_with_message.Message.Text,
	}
}

func createUpdateUrl(token string, last int) string {
	apiUrl := "https://api.telegram.org"
	resource := fmt.Sprintf("/bot%s/getUpdates", token)
	params := url.Values{
		"offset":          {strconv.Itoa(last + 1)},
		"allowed_updates": {"[\"message\"]"},
	}

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	u.RawQuery = params.Encode()
	return u.String()
}

func getNewMessages(url string) ([]message, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("sending request failed with err: %s", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %s", response.Status)
	}

	// body, _ := io.ReadAll(response.Body)
	// fmt.Println("Response Body:", string(body))

	updates := &updateResponse{}
	err = json.NewDecoder(response.Body).Decode(updates)
	if err != nil {
		return nil, fmt.Errorf("decoding failed with error: %s", err)
	}
	// fmt.Println(updates.Result)
	// fmt.Println("--------------------------")

	messages := make([]message, 0, len(updates.Result))
	for i := 0; i < len(updates.Result); i++ {
		// fmt.Println(extract(updates.Result[i]))
		messages = append(messages, extract(updates.Result[i]))
	}

	return messages, nil

	// if err != nil {
	// 	fmt.Println("Error reading response body:", err)
	// 	return
	// }

}

type Message struct {
	Bot  *Bot
	User struct {
		ChatId int
		UserId int
	}
	Text string
}

func (b *Bot) Run(port uint) {
	lastUpdate := 0
	logger.General.Println("Telegram bot started")

	for {
		time.Sleep(b.updateDelay)
		url := createUpdateUrl(b.token, lastUpdate)
		messages, err := getNewMessages(url)
		if err != nil {
			logger.Error.Println("Error reading response: ", err)
			continue
		}
		if len(messages) == 0 {
			continue
		}
		for i := 0; i < len(messages); i++ {
			if messages[i].UpdateId > lastUpdate {
				lastUpdate = messages[i].UpdateId
			}
			if !strings.HasPrefix(messages[i].Text, "/") {
				continue
			}
			for j := 0; j < len(b.commands); j++ {
				if strings.HasPrefix(messages[i].Text, b.commands[j].Keyword) {
					logger.General.Println("<-" + messages[i].Text)
					text := strings.TrimSpace(strings.TrimPrefix(messages[i].Text, b.commands[j].Keyword))
					b.commands[j].Handler(Message{
						Bot: b,
						User: struct {
							ChatId int
							UserId int
						}{messages[i].Chat, messages[i].From},
						Text: text,
					})
					break
				}
			}
		}

	}
}

func (b *Bot) SendMessage(message string, to int) error {
	apiUrl := "https://api.telegram.org"
	resource := fmt.Sprintf("/bot%s/sendMessage", b.token)
	params := url.Values{
		"chat_id": {strconv.Itoa(to)},
		"text":    {message},
	}

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	u.RawQuery = params.Encode()
	requestUrl := u.String()

	response, err := http.Get(requestUrl)
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
	err = b.SendMessage(string(p), b.adminChatId)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
