package app

import (
	"fmt"
	"TeleBotNotifications/internal/db"
	"TeleBotNotifications/internal/config"
	"TeleBotNotifications/internal/spotify"
	"TeleBotNotifications/internal/telegram"
	"TeleBotNotifications/internal/logger"
)

type Server struct {
	bot telegram.Bot
	spotifyClient *spotify.Client
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
	
	logger.Setup(&s.config.Logger)

	s.bot = telegram.NewBot(&s.config.Telegram)

	if s.config.Telegram.AdminChatId != -1 {
		logger.SetupTelegramLogger(func (line string) error {
			return s.bot.SendMessage(line, s.config.Telegram.AdminChatId)
		})
	}

	s.db = db.NewDB(fmt.Sprintf("%s/save.json", s.config.WorkingDirectory))

	s.spotifyClient, err = spotify.NewClient(&s.config.Spotify)
	if err != nil {
		return nil, err
	}
	
	logger.Println("server created")
	return s, nil
}

func (s *Server) Run() {
	s.bot.AddCommand("auth", s.GetCodeFromUrl)
	s.bot.AddCommand("start", s.Greet)
	
	go s.bot.Run(s.config.Port)
	s.db.Load()
	
	s.CheckNewReleases()
}
