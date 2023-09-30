package main

import (
	"TeleBotNotifications/internal/app"
)


func main() {
	server, _ := app.New()
	server.Run()
}
