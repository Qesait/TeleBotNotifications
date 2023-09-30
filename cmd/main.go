package main

import (
	"TeleBotNotifications/internal/app"
)


func main() {
	server, err := app.New()
	if err != nil {
		panic(err)
	}

	server.Run()
}
