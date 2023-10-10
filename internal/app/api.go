package app

import (
	"fmt"
	"net/url"

	"TeleBotNotifications/internal/db"
	"TeleBotNotifications/internal/logger"
	"TeleBotNotifications/internal/telegram"
	"time"
)

func (s *Server) Greet(message telegram.Message) {
	authUrl, err := s.spotifyClient.GenerateAuthUrl()
	if err != nil {
		logger.Error("error generating auth url: ", err)
		return
	}
	text := fmt.Sprintf("Open this link %s.\nCopy and past here url after redirect", *authUrl)
	message.Bot.SendMessage(text, message.User.ChatId)
}

func (s *Server) GetCodeFromUrl(message telegram.Message) {
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

	token, err := s.spotifyClient.RequestAccessToken(&code)
	if err != nil {
		logger.Error("error requesting token: ", err)
		return
	}

	user := db.User{
		UserId:    message.User.UserId,
		ChatId:    message.User.ChatId,
		Token:     *token,
		LastCheck: (time.Now().Add(-120 * time.Hour)).Format("2006-01-02 15:04 -0700 MST"),
	}

	s.db.AddUser(user)

	logger.Println(fmt.Sprintf("New user %d added", user.UserId))
}

func (s *Server) CheckNewReleases () {
	for {
		// TODO: Change update rate
		user := s.db.NextUser()
		if user == nil {
			time.Sleep(time.Minute)
			continue
		}
		logger.Println(fmt.Sprintf("Checking for new releases for user %d", user.UserId))
		LastCheck, err := time.Parse("2006-01-02 15:04 -0700 MST", user.LastCheck)
		if err != nil {
			logger.Error("error parsing time ", err)
		}

		artists, err := s.spotifyClient.GetFollowedArtists(&user.Token)
		if err != nil {
			logger.Error("error getting artists: ", err)
		}

		for _, artist:= range artists {
			lastAlbums, err := s.spotifyClient.GetArtistAlbums(&user.Token, &artist)
			if err != nil {
				logger.Error(fmt.Sprintf("error getting albums for artist %s(%s)", artist.Name, artist.Id), err)
				break
			}
			for _, album := range lastAlbums {
				if LastCheck.Before(album.ReleaseDate) {
					logger.Println(fmt.Sprintf("New release '%s' from %s", album.Name, artist.Name))
					message := album.Url
					s.bot.SendMessage(message, user.ChatId)
				}
			}
			// TODO: Do somethig with this delay
			time.Sleep(2 * time.Second)
		}
		logger.Println(fmt.Sprintf("Finished checking for new releases for user %d", user.UserId))
		time.Sleep(24 * time.Hour)
	}
}