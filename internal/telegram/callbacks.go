package telegram

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type callback struct {
	Keyword string
	Handler CallbackHandler
}

// TODO: type CallbackHandler func(Callback) error
type CallbackHandler func(Callback)

func (b *Bot) AddCallback(keyword string, handler CallbackHandler) {
	b.callbacks = append(b.callbacks, callback{
		Keyword: "/" + keyword,
		Handler: handler,
	})
}

type Callback struct {
	UserId int
	ChatId string
	Data   string
}

func (b *Bot) handleCallback(c *callbackQuery) {
	if c.Data == nil {
		return
	}
	for _, callback := range b.callbacks {
		if strings.HasPrefix(*c.Data, callback.Keyword) {
			// TODO: error checks
			callback.Handler(Callback{
				UserId: c.From.Id,
				ChatId: c.ChatInstance,
				Data:   strings.TrimSpace(strings.TrimPrefix(*c.Data, callback.Keyword)),
			})
			b.answerCallbackQuery(c.Id)
			return
		}
	}
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
