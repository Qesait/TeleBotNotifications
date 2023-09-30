package app

import (
	"fmt"
	"os"
	"log"

	"TeleBotNotifications/internal/db"
	"TeleBotNotifications/internal/api"
	"TeleBotNotifications/pkg/spotify"
	"TeleBotNotifications/pkg/telegram"
)

type Server struct {
	bot telegram.Bot
	spotify_client *spotify.Client
	db db.DB
	// config
}

func New() (*Server, error) {
	var err error

	server := &Server{}
	server.db = db.NewDB("/var/lib/spotify_notifications_bot/save.json")
	server.db.Load()
	log.Println("db loaded")

	var client_id = os.Getenv("SPOTIFY_CLIENT_ID")
	var client_secret = os.Getenv("SPOTIFY_CLIENT_SECRET")
	var scope = "user-follow-read"
	var redirect_uri = "http://localhost:8888"
	
	if client_id == "" || client_secret == "" {
		return nil, fmt.Errorf("no client credentials")
	}

	server.spotify_client, err = spotify.NewClient(client_id, client_secret, redirect_uri, scope)
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
	
	s.bot.Run(8888)
}