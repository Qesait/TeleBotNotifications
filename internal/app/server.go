package app

import (
	"fmt"
	"time"

	"TeleBotNotifications/internal/db"
	"TeleBotNotifications/internal/api"
	"TeleBotNotifications/internal/config"
	"TeleBotNotifications/internal/spotify"
	"TeleBotNotifications/internal/telegram"
	"TeleBotNotifications/internal/logger"
)

type Server struct {
	bot telegram.Bot
	spotify_client *spotify.Client
	db db.DB
	config *config.Config
}

func New() (*Server, error) {
	var err error
	s := &Server{}

	s.config, err = config.NewConfig()
	if err != nil {
		return nil, err
	}

	s.bot = telegram.NewBot(s.config.Telegram)

	if s.config.Telegram.AdminChatId != -1 {
		logger.SetupTelegramLogger(func (line string) error {
			return s.bot.SendMessage(line, s.config.Telegram.AdminChatId)
		})
	}
	logger.TelegramLogLevel = 1
	logger.FileLogLevel = 2
	logger.StdLogLevel = 2

	s.db = db.NewDB("/var/lib/spotify_notifications_bot/save.json")

	s.spotify_client, err = spotify.NewClient(s.config.Spotify)
	if err != nil {
		return nil, err
	}
	
	return s, nil
}

func (s *Server) Run() {
	logger.Println(fmt.Sprintf("%#v", s.config))

	s.db.Load()

	s.bot.AddCommand("auth", api.GetCodeFromUrl(&s.db, s.spotify_client))
	s.bot.AddCommand("start", api.Greet(s.spotify_client))
	
	go s.bot.Run(s.config.Port)
	
	s.CheckNewReleases()
}

func (s *Server) CheckNewReleases () {
	for {
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

		artists, err := s.spotify_client.GetFollowedArtists(&user.Token)
		if err != nil {
			logger.Error("error getting artists: ", err)
		}

		for _, artist:= range artists {
			lastAlbums, err := s.spotify_client.GetArtistAlbums(&user.Token, &artist)
			if err != nil {
				logger.Error(fmt.Sprintf("error getting albums for artist %s(%s)", artist.Name, artist.Id), err)
				break
			}
			for _, album := range lastAlbums {
				if LastCheck.Before(album.ReleaseDate) {
					logger.Println(fmt.Sprintf("New release '%s' from %s\n%s", album.Name, artist.Name, album.Url))
					message := album.Url
					s.bot.SendMessage(message, user.ChatId)
				}
			}
			time.Sleep(2 * time.Second)
		}
		logger.Println(fmt.Sprintf("Finished checking for new releases for user %d", user.UserId))
		time.Sleep(24 * time.Hour)
	}
}