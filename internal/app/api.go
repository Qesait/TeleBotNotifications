package app

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"TeleBotNotifications/internal/db"
	"TeleBotNotifications/internal/logger"
	"TeleBotNotifications/internal/telegram"
)

func (s *Server) Greet(message telegram.ReceivedMessage) {
	authUrl, err := s.spotifyClient.GenerateAuthUrl()
	if err != nil {
		logger.Error.Println("error generating auth url: ", err)
		return
	}
	text := "Press button below to start authentication. Then use \"/auth <URL>\" with URL you were redirected"
	err = s.bot.SendMessage(telegram.BotMessage{
		Text:        text,
		ReplyMarkup: telegram.ButtonRow(telegram.URLButton("Authenticate", *authUrl))})
	if err != nil {
		logger.Error.Println("error sending auth url: ", err)
		return
	}
}

func (s *Server) GetCodeFromUrl(message telegram.ReceivedMessage) {
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
		LastCheck: time.Now(),
	}

	s.db.Set(user)
	err = s.db.Save()
	if err != nil {
		logger.Error.Println("db save failed:", err)
	}

	text := "Successfull authentication"
	err = s.bot.SendMessage(telegram.BotMessage{
		Text: text,
	})
	if err != nil {
		logger.Error.Println("error sending auth response: ", err)
		return
	}
	logger.General.Printf("New user %d added\n", user.UserId)
}

func (s *Server) ForceCheck(message telegram.ReceivedMessage) {
	var days int
	var err error
	if message.Text == "" {
		// TODO: put into config
		days = 7
	} else {
		days, err = strconv.Atoi(strings.TrimSpace(message.Text))
	}
	if err != nil {
		err = s.bot.SendMessage(telegram.BotMessage{
			Text: "Wrong command parameter. It must be a number",
		})
		if err != nil {
			logger.Error.Println("error sending auth response: ", err)
		}
		return
	}

	if s.cancelSpotifyCheck != nil {
		s.cancelSpotifyCheck()
	}

	offset := time.Duration(days) * 24 * time.Hour
	s.CheckNewReleases(&offset)
}

func stripTime(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func (s *Server) CheckNewReleases(offset *time.Duration) {
	user := s.db.Get()
	checkStart := time.Now()
	checkStartDate := stripTime(checkStart)
	var rangeStart time.Time
	if offset == nil {
		rangeStart = user.LastCheck
	} else {
		rangeStart = checkStart.Add(*offset)
	}
	rangeStartDate := stripTime(rangeStart)
	if !checkStartDate.Before(rangeStartDate) {
		return
	}

	// Create context for premature stop
	spotifyContext, cancel := context.WithCancel(context.Background())
	s.cancelSpotifyCheck = cancel

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		user.LastCheck = checkStart
		s.db.Set(*user)
		logger.General.Printf("Checking for new releases. From %s to %s\n", rangeStartDate.Format("2006-01-02"), checkStartDate.Format("2006-01-02"))

		newAlbums, err := s.spotifyClient.GetNewReleases(user.Token, rangeStartDate, checkStartDate, spotifyContext)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				// TODO: maybe print something in general log before death
				return
			}
			logger.Error.Println("Failed to get new releases with error:", err)
			return
		}
		for _, album := range newAlbums {
			select {
			case <-spotifyContext.Done():
				return
			default:
				// TODO: show all artist, or verify that first is main
				logger.General.Printf("\x1b[34mNew release '%s'\tby %s\tfrom %s\n\x1b[0m", album.Name, album.Artists[0].Name, album.ReleaseDate.Format("02.01.2006"))
				parseMode := "Markdown"
				message := telegram.BotMessage{
					Text:        fmt.Sprintf("*%s* · %s[ㅤ](%s)", escapeCharacters(album.Name), escapeCharacters(album.Artists[0].Name), album.Url),
					ParseMode:   &parseMode,
					ReplyMarkup: telegram.ButtonRow(telegram.CallbackButton("Play", "/play "+album.Uri), telegram.CallbackButton("Add to queue", "/queue "+album.Id)),
				}
				// TODO: async sending messages
				err := s.bot.SendMessage(message)
				if err != nil {
					logger.Error.Println("error sending message with new release:", err)
				}
			}
		}
		logger.General.Println("Finished checking for new releases")
		err = s.db.Save()
		if err != nil {
			logger.Error.Println("db save failed:", err)
		}
	}()
}

func escapeCharacters(raw string) string {
	new := ""
	for _, ch := range raw {
		if ch == '_' || ch == '*' || ch == '`' || ch == '[' {
			new = new + "\\"
		}
		new = new + string(ch)
	}
	return new
}

func (s *Server) AddToQueue(callback telegram.Callback) {
	user := s.db.Get()
	if user == nil {
		logger.Error.Println("No spotify account authorized")
		return
	}
	tracks, err := s.spotifyClient.GetAlbumTracks(&user.Token, callback.Data, 50, 0, nil)
	if err != nil {
		logger.Error.Printf("failed getting album tracks: %s\n", err)
	}
	for _, track := range tracks {
		err = s.spotifyClient.AddItemtoPlaybackQueue(&user.Token, &track.Uri, nil)
		if err != nil {
			logger.Error.Printf("add to queue failed with error: %s\n", err)
			// TODO: collect errors or skip them. Maybe problem with one track only, but maybe I will get 50 notifications for album
			break
		}
	}
}

func (s *Server) PlayTrack(callback telegram.Callback) {
	user := s.db.Get()
	if user == nil {
		logger.Error.Println("No spotify account authorized")
		return
	}
	err := s.spotifyClient.StartResumePlayback(&user.Token, &callback.Data, nil)
	if err != nil {
		logger.Error.Printf("play track failed with error: %s\n", err)
	}
}
