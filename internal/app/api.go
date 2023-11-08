package app

import (
	"fmt"
	"net/url"

	"TeleBotNotifications/internal/db"
	"TeleBotNotifications/internal/logger"
	"TeleBotNotifications/internal/telegram"
	"time"
)

func (s *Server) Greet(message telegram.ReceivedMessage) {
	authUrl, err := s.spotifyClient.GenerateAuthUrl()
	if err != nil {
		logger.Error.Println("error generating auth url: ", err)
		return
	}
	text := fmt.Sprintf("Open this link %s.\nCopy and past here url after redirect", *authUrl)
	s.bot.SendMessage(telegram.BotMessage{ChatId: message.ChatId, Text: text})
}

func (s *Server) GetCodeFromUrl(message telegram.ReceivedMessage) {
	parsedURL, err := url.Parse(message.Text)
	if err != nil {
		logger.Error.Println("error parsing URL: ", err)
		return
	}
	code := parsedURL.Query().Get("code")
	if code == "" {
		logger.Error.Println("couldn't extract code from url: ", message.Text)
		return
	}

	token, err := s.spotifyClient.RequestAccessToken(&code)
	if err != nil {
		logger.Error.Println("error requesting token: ", err)
		return
	}

	user := db.User{
		UserId:    message.UserId,
		ChatId:    message.ChatId,
		Token:     *token,
		LastCheck: (time.Now().Add(-120 * time.Hour)).Format("2006-01-02 15:04 -0700 MST"),
	}

	s.db.AddUser(user)

	logger.General.Printf("New user %d added\n", user.UserId)
}

func stripTime(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func (s *Server) CheckNewReleases() {
	for {
		user := s.db.NextUser()
		if user == nil {
			time.Sleep(time.Minute)
			continue
		}
		logger.General.Printf("Checking for new releases for user %d. ", user.UserId)
		lastCheck, err := time.Parse("2006-01-02 15:04 -0700 MST", user.LastCheck)
		if err != nil {
			logger.Error.Println("error parsing time ", err)
		}
		logger.General.Printf("Previous check was: %s\n", lastCheck)

		// TODO: somehow aacount for new users
		currentTime := stripTime(time.Now())
		lastCheck = stripTime(lastCheck)
		if lastCheck == currentTime {
			logger.General.Printf("Already checked new releases for user %d today", user.UserId)
			time.Sleep(time.Hour)
			continue
		}

		artists, err := s.spotifyClient.GetFollowedArtists(&user.Token)
		if err != nil {
			logger.Error.Println("error getting artists: ", err)
		}
		logger.General.Println("Going to check", len(artists), "artists")

		for _, artist := range artists {
			lastAlbums, err := s.spotifyClient.GetArtistAlbums(&user.Token, &artist)
			if err != nil {
				logger.Error.Printf("error getting albums for artist %s(%s): %s\n", artist.Name, artist.Id, err)
				break
			}
			for _, album := range lastAlbums {
				if !lastCheck.After(album.ReleaseDate) && currentTime.After(album.ReleaseDate) {
					logger.General.Printf("\x1b[34mNew release '%s'\tby %s\tfrom %s\n\x1b[0m", album.Name, artist.Name, album.ReleaseDate.Format("02.01.2006"))
					parseMode := "Markdown"
					replyMarkup := "{\"inline_keyboard\": [[{\"text\": \"Add to queue\",\"callback_data\": \"queue\"}]]}"
					err := s.bot.SendMessage(telegram.BotMessage{
						ChatId:      user.ChatId,
						Text:        fmt.Sprintf("*%s* · %s[ㅤ](%s)", escapeCharacters(album.Name), escapeCharacters(album.Artists[0].Name), album.Url),
						ParseMode:   &parseMode,
						ReplyMarkup: &replyMarkup})
					if err != nil {
						logger.Error.Println("error sending message with new release:", err)
					}
				}
			}
			// TODO: Do somethig with this delay
			time.Sleep(1 * time.Second)
		}
		logger.General.Printf("Finished checking for new releases for user %d", user.UserId)
		time.Sleep(24 * time.Hour)
	}
}

func escapeCharacters(raw string) string {
	new := ""
	for _, ch := range raw {
		if ch == '_' || ch == '*' || ch == '`' || ch == '[' {
			new = new + "\\"
		}
		new = new + string(ch)
	}
	return new
}
