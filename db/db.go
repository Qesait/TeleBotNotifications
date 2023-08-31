package db

import (
	"TeleBotNotifications/spotify"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type User struct {
	UserId int                 `json:"user_id"`
	ChatId int                 `json:"chat_id"`
	Token  spotify.OAuth2Token `json:"token"`
	Artist []spotify.Artist    `json:"artists"`
}

type dB struct {
	users    []User
	saveFile string
	nextUser uint
}

func NewDB(saveFile string) dB {
	return dB{saveFile: saveFile, nextUser: 0}
}

func (db *dB) Load() {
	jsonFile, err := os.Open(db.saveFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &db.users)
}

func (db *dB) save() {
	jsonFile, err := os.Open(db.saveFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer jsonFile.Close()

	byteValue, err := json.Marshal(db.users)

	if err != nil {
		fmt.Println(err)
		return
	}
	jsonFile.Write(byteValue)
}

func (db *dB) AddUser(user User) {
	db.users = append(db.users, user)
	db.save()
}

func (db *dB) NextUser() User {
	user := db.users[db.nextUser]
	db.nextUser += 1
	return user
}
