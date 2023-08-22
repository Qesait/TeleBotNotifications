package telegram

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type CommandHandler func(string) (*string, error)

type command struct {
	Keyword string
	Handler CommandHandler
}

type Bot struct {
	token       string
	commands    []command
	http_client *http.Client
}

func NewBot(token string) Bot {
	return Bot{token: token, http_client: &http.Client{}}
}

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

func (b *Bot) Go(port uint) {
	lastUpdate := 0

	for {
		time.Sleep(time.Second * 5)
		url := createUpdateUrl(b.token, lastUpdate)
		messages, err := getNewMessages(url)
		if err != nil {
			fmt.Println("Error reading response", err)
			continue
		}
		if len(messages) == 0 {
			fmt.Println("Nothing new")
			continue
		}
		for i := 0; i < len(messages); i++ {
			// fmt.Println(messages[i].Text)
			if messages[i].UpdateId > lastUpdate {
				lastUpdate = messages[i].UpdateId
			}
			if !strings.HasPrefix(messages[i].Text, "/") {
				continue
			}
			for j := 0; j < len(b.commands); j++ {
				// fmt.Println(messages[i].Text, b.commands[j].Keyword)
				if strings.HasPrefix(messages[i].Text, b.commands[j].Keyword) {
					text:= strings.TrimSpace(strings.TrimPrefix(messages[i].Text, b.commands[j].Keyword))
					fmt.Println("<-", text)
					response, err := b.commands[j].Handler(text)
					if err != nil {
						fmt.Printf("Command \"%s\" failed with error: %s\n", b.commands[j].Keyword, err)
						b.SendMessage("Something went wrong(", messages[i].From)	
					} else {
						fmt.Println("->", *response)
						b.SendMessage(*response, messages[i].From)
					}
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
