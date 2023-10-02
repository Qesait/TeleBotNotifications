package app

import (
	"fmt"
	"os"
	"log"

	"TeleBotNotifications/internal/db"
	"TeleBotNotifications/internal/api"
	"TeleBotNotifications/internal/config"
	"TeleBotNotifications/internal/spotify"
	"TeleBotNotifications/internal/telegram"
)

type Server struct {
	bot telegram.Bot
	spotify_client *spotify.Client
	db db.DB
	config *config.Config
}

func New() (*Server, error) {
	var err error
	server := &Server{}

	server.config, err = config.NewConfig()
	if err != nil {
		return nil, err
	}

	server.db = db.NewDB("/var/lib/spotify_notifications_bot/save.json")
	server.db.Load()
	log.Println("db loaded")

	server.spotify_client, err = spotify.NewClient(server.config.Spotify)
	if err != nil {
		return nil, err
	}
	
	bot_token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if bot_token == "" {
		return nil, fmt.Errorf("no bot credentials")
	}
	server.bot = telegram.NewBot(bot_token)

	return server, nil
}

func (s *Server) Run() {
	s.bot.AddCommand("auth", api.GetCodeFromUrl(&s.db, s.spotify_client))
	s.bot.AddCommand("start", api.Greet(s.spotify_client))
	
	go api.CheckNewReleases(&s.db, s.spotify_client, &s.bot)
	
	s.bot.Run(s.config.Port)
}