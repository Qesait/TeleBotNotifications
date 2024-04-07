package app

import (
	"context"
	"fmt"
	"errors"
	"net/url"
	"time"

	"TeleBotNotifications/internal/db"
	"TeleBotNotifications/internal/logger"
	"TeleBotNotifications/internal/spotify"
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

func stripTime(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func (s *Server) CheckNewReleasesAlbum(album spotify.Album, token spotify.OAuth2Token, rangeStart, rangeEnd time.Time, ctx context.Context) {
	if rangeStart.After(album.ReleaseDate) || rangeEnd.Before(album.ReleaseDate) {
		return
	}
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

func (s *Server) CheckNewReleasesArtist(artist spotify.Artist, token spotify.OAuth2Token, rangeStart, rangeEnd time.Time, ctx context.Context) error {
	lastAlbums, err := s.spotifyClient.GetArtistAlbums(&token, &artist)
	if err != nil {
		return fmt.Errorf("error getting albums for artist %s(%s): %s", artist.Name, artist.Id, err)
	}
	for _, album := range lastAlbums {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
			s.CheckNewReleasesAlbum(album, token, rangeStart, rangeEnd, ctx)
		}
	}
	return nil
}

func (s *Server) CheckNewReleases(token spotify.OAuth2Token, rangeStart, rangeEnd time.Time, ctx context.Context) error {
	artists, err := s.spotifyClient.GetFollowedArtists(&token)
	if err != nil {
		return fmt.Errorf("error getting artists: %w", err)
	}
	logger.General.Println("Going to check", len(artists), "artists")

	failedChecks := 0
	for _, artist := range artists {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
			// TODO: Do somethig with this delay
			if failedChecks > 10 {
				break
			}
			err := s.CheckNewReleasesArtist(artist, token, rangeStart, rangeEnd, ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return err
				}
				if failedChecks == 0 {
					logger.Error.Printf("failed checking new releases for artist %s(%s): %s\n", artist.Name, artist.Id, err)
				}
				failedChecks += 1
				continue

			}
			time.Sleep(1 * time.Second)
		}
	}
	if failedChecks > 0 {
		return fmt.Errorf("Check for new releases failed for %d artists\n", failedChecks)
	}
	return nil
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
