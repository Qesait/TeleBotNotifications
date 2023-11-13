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
	s.bot.AddCommand("start", "Get a link to steal your account", s.Greet)
	s.bot.AddCallback("queue", s.AddToQueue)
	s.bot.AddCallback("play", s.PlayTrack)

	go s.bot.Run(s.config.Port)
	s.db.Load()

	s.CheckNewReleases()
}

func (s *Server) WriteQuery(callback telegram.Callback) {
	s.bot.SendMessage(telegram.BotMessage{ChatId: callback.UserId, Text: callback.Data})
}

func (s *Server) AddToQueue(callback telegram.Callback) {
	user, err := s.db.Get(callback.UserId)
	logger.General.Println("Hi from add to queue")
	if err != nil {
		logger.Error.Printf("add to queue failed with error: %s\n", err)
	}
	err = s.spotifyClient.AddItemtoPlaybackQueue(&user.Token, &callback.Data, nil)
	if err != nil {
		logger.Error.Printf("add to queue failed with error: %s\n", err)
	}
}

func (s *Server) PlayTrack(callback telegram.Callback) {
	user, err := s.db.Get(callback.UserId)
	if err != nil {
		logger.Error.Printf("play track failed with error: %s\n", err)
	}
	err = s.spotifyClient.StartResumePlayback(&user.Token, &callback.Data, nil)
	if err != nil {
		logger.Error.Printf("play track failed with error: %s\n", err)
	}
}