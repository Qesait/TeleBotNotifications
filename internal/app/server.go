package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"TeleBotNotifications/internal/config"
	"TeleBotNotifications/internal/db"
	"TeleBotNotifications/internal/logger"
	"TeleBotNotifications/internal/spotify"
	"TeleBotNotifications/internal/telegram"
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
	err := s.db.Load()
	if err != nil {
		logger.Error.Println("db load error", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// Handle termination signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Start goroutines
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.bot.AddCommand("auth", "submit an authentication link", s.GetCodeFromUrl)
		s.bot.AddCommand("start", "Get a link to steal your account", s.Greet)
		s.bot.AddCallback("queue", s.AddToQueue)
		s.bot.AddCallback("play", s.PlayTrack)

		s.bot.Run(s.config.Port, ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.CheckNewReleases(ctx)
	}()

	// Wait for termination signal
	<-sigs
	logger.General.Println("Shutdown signal received, exiting...")
	cancel() // Send cancellation signal to goroutines

	wg.Wait() // Wait for all goroutines to finish
	logger.General.Println("Application shutdown gracefully")
}
