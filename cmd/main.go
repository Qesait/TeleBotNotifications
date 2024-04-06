// package main

// import (
// 	"TeleBotNotifications/internal/app"
// 	"time"
// )

// func main() {
// 	loc, _ := time.LoadLocation("UTC")
//     time.Local = loc

// 	server, err := app.New()
// 	if err != nil {
// 		panic(err)
// 	}

// 	server.Run()
// }

package main

import (
	"TeleBotNotifications/internal/config"
	"log"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(cfg)
}