package main

import (
	"TeleBotNotifications/internal/app"
	"time"
)

func main() {
	loc, _ := time.LoadLocation("UTC")
    time.Local = loc

	server, err := app.New()
	if err != nil {
		panic(err)
	}

	server.Run()
}
