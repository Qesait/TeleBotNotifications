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
		logger.Error.Println("error generating auth url: ", err)
		return
	}
	text := fmt.Sprintf("Open this link %s.\nCopy and past here url after redirect", *authUrl)
	s.bot.SendMessage(text, message.ChatId)
}

func (s *Server) GetCodeFromUrl(message telegram.Message) {
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

func (s *Server) CheckNewReleases () {
	for {
		// TODO: Change update rate
		user := s.db.NextUser()
		if user == nil {
			time.Sleep(time.Minute)
			continue
		}
		logger.General.Printf("Checking for new releases for user %d. Previous check was: %s\n", user.UserId, user.LastCheck)
		LastCheck, err := time.Parse("2006-01-02 15:04 -0700 MST", user.LastCheck)
		if err != nil {
			logger.Error.Println("error parsing time ", err)
		}

		artists, err := s.spotifyClient.GetFollowedArtists(&user.Token)
		if err != nil {
			logger.Error.Println("error getting artists: ", err)
		}

		for _, artist:= range artists {
			lastAlbums, err := s.spotifyClient.GetArtistAlbums(&user.Token, &artist)
			if err != nil {
				logger.Error.Printf("error getting albums for artist %s(%s): %s\n", artist.Name, artist.Id, err)
				break
			}
			for _, album := range lastAlbums {
				if LastCheck.Before(album.ReleaseDate) {
					logger.General.Printf("New release '%s'\tby %s\tfrom %s\n", album.Name, artist.Name, album.ReleaseDate.Format("02.01.2006"))
					message := album.Url
					s.bot.SendMessage(message, user.ChatId)
				}
			}
			// TODO: Do somethig with this delay
			time.Sleep(2 * time.Second)
		}
		logger.General.Printf("Finished checking for new releases for user %d", user.UserId)
		time.Sleep(24 * time.Hour)
	}
}