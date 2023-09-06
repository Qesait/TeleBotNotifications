package main

import (
	// "errors"
	// "fmt"
	// "net/http"
	"fmt"
	"net/url"
	"os"

	"TeleBotNotifications/db"
	"TeleBotNotifications/spotify"
	"TeleBotNotifications/telegram"
	"time"
)

// func server(port uint, c chan string) {
// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 		// fmt.Println(r)
// 		fmt.Println(r.URL.Query().Get("idk"))
// 		c <- r.URL.Query().Get("code")
// 	})
// 	server := http.Server{
// 		Addr:    fmt.Sprintf(":%d", port),
// 		Handler: mux,
// 	}
// 	if err := server.ListenAndServe(); err != nil {
// 		if !errors.Is(err, http.ErrServerClosed) {
// 			fmt.Printf("error running http server: %s\n", err)
// 		}
// 	}
// }
// const serverPort = 8888
// var c = make(chan string, 1)
// var authorization_code = ""


func Greet(spotify_client *spotify.Client) func (telegram.Message) {
	greet := func (message telegram.Message) {
		authUrl, err := spotify_client.GenerateAuthUrl()
		if err != nil {
			message.Bot.SendMessage(fmt.Sprintf("error generating auth url: %s", err), message.User.ChatId)
			fmt.Println("error generating auth url: ", err)
			return
		}
		message.Bot.SendMessage(*authUrl, message.User.ChatId)
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
		// artists, err := spotify_client.GetFollowedArtists(token)
		// if err != nil {
		// 	message.Bot.SendMessage(fmt.Sprintf("error getting artists: %s", err), message.User.ChatId)
		// 	fmt.Println("error getting artists: ", err)
		// 	return
		// }

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

// func GetInfo(dB *db.DB) func (telegram.Message) {
// 	getInfo := func(message telegram.Message) {
// 		message.Bot.SendMessage(fmt.Sprintf("%d", len(dB.NextUser().Artist)), message.User.ChatId)
// 		fmt.Println(dB.NextUser())
// 	}
// 	return getInfo
// }

func CheckNewReleases(dB *db.DB, spotifyClient *spotify.Client, bot *telegram.Bot) {
	for {
		time.Sleep(1 * time.Minute)

		user := dB.NextUser()
		bot.SendMessage("Checking for new releases", user.ChatId)
		isSomethingNew := false
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
			fmt.Println("Checking", artist.Name, artist.Id)
			for _, album := range lastAlbums {
				if LastCheck.Before(album.ReleaseDate) {
					isSomethingNew = true
					message := fmt.Sprintf("New release '%s' from %s", album.Name, artist.Name)
					bot.SendMessage(message, user.ChatId)
					fmt.Println(message, LastCheck.String(), album.ReleaseDate.String())
				}
			}
			time.Sleep(3 * time.Second)
		}
		if !isSomethingNew {
			bot.SendMessage("Did not found new releases", user.ChatId)
			fmt.Println("Did not found new releases")
		}
	}
}

func main() {
	// dB := db.NewDB("/var/lib/spotify_notifications_bot/save.json")
	dB := db.NewDB("C:\\Users\\Toolen\\go\\src\\TeleBotNotifications\\save.json")
	dB.Load()
	fmt.Println("db loaded")

	var client_id = os.Getenv("spotify_client_id")
	var client_secret = os.Getenv("spotify_client_secret")
	var scope = "user-follow-read"
	var redirect_uri = "http://localhost:8888"
	
	if client_id == "" || client_secret == "" {
		panic("no client credentials")
	}

	spotify_client, _ := spotify.NewClient(client_id, client_secret, redirect_uri, scope)
	
	bot_token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if bot_token == "" {
		panic("no bot credentials")
	}
	bot := telegram.NewBot(bot_token)
	bot.AddCommand("auth", GetCodeFromUrl(&dB, spotify_client))
	bot.AddCommand("start", Greet(spotify_client))
	// bot.AddCommand("info", GetInfo(&dB))

	go CheckNewReleases(&dB, spotify_client, &bot)

	bot.Run(8888)



	// if authorization_code == "" {
	// 	go server(serverPort, c)
	// 	time.Sleep(100 * time.Millisecond)

	// 	var auth_url, _ = spotify_client.GenerateAuthUrl()
	// 	fmt.Println(*auth_url)

	// 	authorization_code = <-c // not access token
	// 	fmt.Println(authorization_code)
	// }

	// token, _ := spotify_client.RequestAccessToken(&authorization_code)
	// fmt.Println(*token)
	// artists, err := spotify_client.GetFollowedArtists(token)
	// fmt.Println("")
	// fmt.Println(err)
	// if err == nil {
	// 	fmt.Println(len(artists))
	// 	for i:=0; i<len(artists); i++ {
	// 		fmt.Println(artists[i])
	// 	}
	// }
}
