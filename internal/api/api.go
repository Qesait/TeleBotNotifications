package api

import (
	"fmt"
	"net/url"

	"TeleBotNotifications/internal/db"
	"TeleBotNotifications/internal/spotify"
	"TeleBotNotifications/internal/telegram"
	"time"
)



func Greet(spotify_client *spotify.Client) func (telegram.Message) {
	greet := func (message telegram.Message) {
		authUrl, err := spotify_client.GenerateAuthUrl()
		if err != nil {
			message.Bot.SendMessage(fmt.Sprintf("error generating auth url: %s", err), message.User.ChatId)
			fmt.Println("error generating auth url: ", err)
			return
		}
		text := fmt.Sprintf("Open this link %s.\nCopy and past here url after redirect", *authUrl)
		message.Bot.SendMessage(text, message.User.ChatId)
	}
	return greet
}


func GetCodeFromUrl(dB *db.DB, spotify_client *spotify.Client) func(telegram.Message) {
	getCodeFromUrl := func(message telegram.Message) {
		parsedURL, err := url.Parse(message.Text)
		if err != nil {
			message.Bot.SendMessage(fmt.Sprintf("error parsing URL: %s", err), message.User.ChatId)
			fmt.Println("error parsing URL: ", err)
			return
		}
		code := parsedURL.Query().Get("code")
		if code == "" {
			message.Bot.SendMessage(fmt.Sprintf("couldn't extract code from url: %s", message.Text), message.User.ChatId)
			fmt.Println("couldn't extract code from url: ", message.Text)
			return
		}

		token, err := spotify_client.RequestAccessToken(&code)
		if err != nil {
			message.Bot.SendMessage(fmt.Sprintf("error requesting token: %s", err), message.User.ChatId)
			fmt.Println("error requesting token: ", err)
			return
		}

		user := db.User{
			UserId: message.User.UserId,
			ChatId: message.User.ChatId,
			Token: *token,
			LastCheck: (time.Now().Add(-120 * time.Hour)).Format("2006-01-02 15:04 -0700 MST"),
		}

		fmt.Println("got user:", user)

		dB.AddUser(user)

		message.Bot.SendMessage("Looks like it worked", message.User.ChatId)
		fmt.Println(user.LastCheck)
		
	}

	return getCodeFromUrl
}

func CheckNewReleases(dB *db.DB, spotifyClient *spotify.Client, bot *telegram.Bot) {
	time.Sleep(15 * time.Second)
	for {
		user := dB.NextUser()
		if user == nil {
			time.Sleep(time.Minute)
			continue
		}
		// bot.SendMessage("Checking for new releases", user.ChatId)
		// isSomethingNew := false
		LastCheck, err := time.Parse("2006-01-02 15:04 -0700 MST", user.LastCheck)
		if err != nil {
			fmt.Println("error parsing time ", err)
		}

		artists, err := spotifyClient.GetFollowedArtists(&user.Token)
		if err != nil {
			bot.SendMessage(fmt.Sprintf("error getting artists: %s", err), user.ChatId)
			fmt.Println("error getting artists: ", err)
			return
		}

		for i, artist:= range artists {
			lastAlbums, err := spotifyClient.GetArtistAlbums(&user.Token, &artist)
			if err != nil {
				bot.SendMessage(fmt.Sprintf("error getting albums: %s", err), user.ChatId)
				fmt.Println("error getting albums: ", err, artist.Name, artist.Id, i)
				break
			}
			fmt.Printf("Checking %d albums from %s (%s)\n", len(lastAlbums), artist.Name, artist.Id)
			for _, album := range lastAlbums {
				if LastCheck.Before(album.ReleaseDate) {
					// isSomethingNew = true
					// message := fmt.Sprintf("New release '%s' from %s\n%s", album.Name, artist.Name, album.Url)
					message := album.Url
					bot.SendMessage(message, user.ChatId)
					fmt.Println(message, LastCheck.String(), album.ReleaseDate.String())
				}
			}
			time.Sleep(2 * time.Second)
		}
		// if !isSomethingNew {
		// 	bot.SendMessage("Did not found new releases", user.ChatId)
		// 	fmt.Println("Did not found new releases")
		// } else {
		// 	bot.SendMessage("All done for now", user.ChatId)
		// 	fmt.Println("All done for now")
		// }
		time.Sleep(24 * time.Hour)
	}
}