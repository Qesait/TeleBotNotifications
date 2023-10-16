package app

import (
	"TeleBotNotifications/internal/config"
	"TeleBotNotifications/internal/db"
	"TeleBotNotifications/internal/logger"
	"TeleBotNotifications/internal/spotify"
	"TeleBotNotifications/internal/telegram"
	"fmt"
)

type Server struct {
	bot           telegram.Bot
	spotifyClient *spotify.Client
	db            db.DB
	config        *config.Config
}

func New() (*Server, error) {
	var err error
	s := &Server{}

	s.config, err = config.NewConfig()
	if err != nil {
		return nil, err
	}

	s.bot = telegram.NewBot(&s.config.Telegram)
	err = logger.Setup(&s.config.Logger, &s.bot)
	if err != nil {
		return nil, err
	}

	s.db = db.NewDB(fmt.Sprintf("%s/save.json", s.config.WorkingDirectory))

	s.spotifyClient, err = spotify.NewClient(&s.config.Spotify)
	if err != nil {
		return nil, err
	}

	logger.General.Println("server created")
	return s, nil
}

func (s *Server) Run() {
	s.bot.AddCommand("auth", "submit an authentication link", s.GetCodeFromUrl)
	s.bot.AddCommand("start", "Get a link to steal an account", s.Greet)

	go s.bot.Run(s.config.Port)
	s.db.Load()

	s.CheckNewReleases()
}
