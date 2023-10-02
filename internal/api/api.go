package api

import (
	"fmt"
	"net/url"

	"TeleBotNotifications/internal/db"
	"TeleBotNotifications/internal/spotify"
	"TeleBotNotifications/internal/telegram"
	"TeleBotNotifications/internal/logger"
	"time"
)



func Greet(spotify_client *spotify.Client) func (telegram.Message) {
	greet := func (message telegram.Message) {
		authUrl, err := spotify_client.GenerateAuthUrl()
		if err != nil {
			logger.Error("error generating auth url: ", err)
			return
		}
		text := fmt.Sprintf("Open this link %s.\nCopy and past here url after redirect", *authUrl)
		message.Bot.SendMessage(text, message.User.ChatId)
	}

	return greet
}


func GetCodeFromUrl(dB *db.DB, spotify_client *spotify.Client) func (telegram.Message) {
	getCodeFromUrl := func(message telegram.Message) {
		parsedURL, err := url.Parse(message.Text)
		if err != nil {
			logger.Error("error parsing URL: ", err)
			return
		}
		code := parsedURL.Query().Get("code")
		if code == "" {
			logger.Error("couldn't extract code from url: ", fmt.Errorf(message.Text))
			return
		}

		token, err := spotify_client.RequestAccessToken(&code)
		if err != nil {
			logger.Error("error requesting token: ", err)
			return
		}

		user := db.User{
			UserId: message.User.UserId,
			ChatId: message.User.ChatId,
			Token: *token,
			LastCheck: (time.Now().Add(-120 * time.Hour)).Format("2006-01-02 15:04 -0700 MST"),
		}

		dB.AddUser(user)

		logger.Println(fmt.Sprintf("New user %d added", user.UserId))
	}

	return getCodeFromUrl
}
