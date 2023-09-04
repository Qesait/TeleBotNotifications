package main

import (
	// "errors"
	// "fmt"
	// "net/http"
	"fmt"
	"net/url"
	"os"

	// "time"
	// "TeleBotNotifications/spotify"
	"TeleBotNotifications/telegram"
	"TeleBotNotifications/db"
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

func Greet(message telegram.Message) {
	response := "Greetings!"
	message.Bot.SendMessage(response, message.User.ChatId)
}

func GetCodeFromUrl(message telegram.Message) {
	parsedURL, err := url.Parse(message.Text)
	if err != nil {
		message.Bot.SendMessage(fmt.Sprintf("error parsing URL: %s", err), message.User.ChatId)
		fmt.Println("error parsing URL: ", err)
	}
	code := parsedURL.Query().Get("code")
	if code == "" {
		message.Bot.SendMessage(fmt.Sprintf("couldn't extract code from url: %s", message.Text), message.User.ChatId)
		fmt.Println("couldn't extract code from url: ", message.Text)
	}
	message.Bot.SendMessage(code, message.User.ChatId)
}

func main() {
	db := db.NewDB("/var/lib/spotify_notifications_bot/save.json")
	db.Load()


	bot_token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if bot_token == "" {
		panic("no bot credentials")
	}
	bot := telegram.NewBot(bot_token)
	bot.AddCommand("auth", GetCodeFromUrl)
	bot.AddCommand("start", Greet)
	bot.AddCommand("start2", Greet)
	bot.Run(8888)

	// var client_id = os.Getenv("spotify_client_id")
	// var client_secret = os.Getenv("spotify_client_secret")
	// var scope = "user-follow-read"
	// var redirect_uri = "http://localhost:8888"

	// if client_id == "" || client_secret == "" {
	// 	panic("no client credentials")
	// }

	// spotify_client, _ := spotify.NewClient(client_id, client_secret, redirect_uri, scope)

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
