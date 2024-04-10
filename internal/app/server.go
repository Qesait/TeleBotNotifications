package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"TeleBotNotifications/internal/config"
	"TeleBotNotifications/internal/db"
	"TeleBotNotifications/internal/logger"
	"TeleBotNotifications/internal/spotify"
	"TeleBotNotifications/internal/telegram"
)

type Server struct {
	bot                telegram.Bot
	spotifyClient      *spotify.Client
	db                 db.DB
	config             *config.Config
	cancelSpotifyCheck context.CancelFunc
	wg                 sync.WaitGroup
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

	logger.General.Println("Server created")
	return s, nil
}

func (s *Server) Run() {
	logger.General.Println("Starting server")

	err := s.db.Load()
	logger.General.Println("DB loaded")
	if err != nil {
		logger.Error.Println("DB load error", err)
		return
	}

	// For stopping goroutins on exit signal
	generalContext, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	s.bot.AddCommand("auth", "submit an authentication link", s.GetCodeFromUrl)
	s.bot.AddCommand("start", "Get a link to steal your account", s.Greet)
	s.bot.AddCommand("check", "Find new releses in the past n days (default 7)", s.ForceCheck)

	s.bot.AddCallback("queue", s.AddToQueue)
	s.bot.AddCallback("play", s.PlayTrack)

	err = s.bot.UpdateCommands()
	if err != nil {
		logger.Error.Println("can't update telegram commans: ", err)
		return
	}

	tgUpdateSignal := make(chan struct{}, 1)
	tgUpdateSignal <- struct{}{}
	// TODO: make config variable
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	logger.General.Println("Bot started")

Loop:
	for {
		select {
		case <-sigs:
			cancel()
			if s.cancelSpotifyCheck != nil {
				s.cancelSpotifyCheck()
			}
			break Loop
		case <-tgUpdateSignal:
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				err := s.bot.HandleUpdates(generalContext)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return
					}
					logger.Error.Println(err)
					time.Sleep(1 * time.Second)
				}
				select {
				case <-generalContext.Done():
					return
				default:
					tgUpdateSignal <- struct{}{}
				}
			}()
		case <-ticker.C:
			s.CheckNewReleases(nil, false)
		}
	}

	s.wg.Wait()
	logger.Error.Println("Bot stopped")
}
