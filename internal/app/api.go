package app

import (
	"context"
	"fmt"
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
		ChatId:      message.ChatId,
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

	logger.General.Printf("New user %d added\n", user.UserId)
}

func stripTime(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func (s *Server) CheckNewReleases(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Stopping check for new releases...")
			return
		default:
			user := s.db.Get()
			if user == nil {
				time.Sleep(time.Minute)
				continue
			}

			currentTime := stripTime(time.Now())
			lastCheck := stripTime(user.LastCheck)
			if lastCheck == currentTime {
				time.Sleep(time.Hour)
				continue
			}
			logger.General.Printf("Checking for new releases. Previous check was: %s\n", lastCheck)

			artists, err := s.spotifyClient.GetFollowedArtists(&user.Token)
			if err != nil {
				logger.Error.Println("error getting artists: ", err)
				time.Sleep(time.Hour)
				continue
			}
			logger.General.Println("Going to check", len(artists), "artists")

			failedChecks := 0
			for _, artist := range artists {
				if failedChecks > 10 {
					break
				}
				var lastAlbums []spotify.Album
				lastAlbums, err = s.spotifyClient.GetArtistAlbums(&user.Token, &artist)
				if err != nil {
					// TODO: looks bad
					time.Sleep(10 * time.Second)
					lastAlbums, err = s.spotifyClient.GetArtistAlbums(&user.Token, &artist)
					if err != nil {
						if failedChecks == 0 {
							logger.Error.Printf("error getting albums for artist %s(%s): %s\n", artist.Name, artist.Id, err)
						}
						failedChecks += 1
						// TODO: I don't like every sleep here
						time.Sleep(10 * time.Second)
						continue
					}
				}
				for _, album := range lastAlbums {
					if !lastCheck.After(album.ReleaseDate) && currentTime.After(album.ReleaseDate) {
						logger.General.Printf("\x1b[34mNew release '%s'\tby %s\tfrom %s\n\x1b[0m", album.Name, artist.Name, album.ReleaseDate.Format("02.01.2006"))
						parseMode := "Markdown"
						message := telegram.BotMessage{
							ChatId:      user.ChatId,
							Text:        fmt.Sprintf("*%s* · %s[ㅤ](%s)", escapeCharacters(album.Name), escapeCharacters(album.Artists[0].Name), album.Url),
							ParseMode:   &parseMode,
							ReplyMarkup: telegram.ButtonRow(telegram.CallbackButton("Play", "/play "+album.Uri), telegram.CallbackButton("Add to queue", "/queue "+album.Id)),
						}
						err := s.bot.SendMessage(message)
						if err != nil {
							time.Sleep(10 * time.Second)
							logger.Error.Println("error sending message with new release:", err)
						}
					}
				}
				// TODO: Do somethig with this delay
				time.Sleep(1 * time.Second)
			}
			if failedChecks > 0 {
				logger.Error.Printf("Finished checking for new releases with %d fails\n", failedChecks)
			} else {
				logger.General.Println("Finished checking for new releases")
				user.LastCheck = currentTime
				s.db.Set(*user)
				err := s.db.Save()
				if err != nil {
					logger.Error.Println("db save failed:", err)
				}
			}
			time.Sleep(time.Hour)
		}
	}
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
	if user != nil {
		logger.Error.Println("No spotify account authorized")
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
	if user != nil {
		logger.Error.Println("No spotify account authorized")
	}
	err := s.spotifyClient.StartResumePlayback(&user.Token, &callback.Data, nil)
	if err != nil {
		logger.Error.Printf("play track failed with error: %s\n", err)
	}
}
