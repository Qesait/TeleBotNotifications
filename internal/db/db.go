package db

import (
	"TeleBotNotifications/internal/logger"
	"TeleBotNotifications/internal/spotify"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

type User struct {
	UserId    int                 `json:"user_id"`
	ChatId    int                 `json:"chat_id"`
	Token     spotify.OAuth2Token `json:"token"`
	LastCheck string              `json:"last_check"`
}

type DB struct {
	users    []User
	saveFile string
	nextUser int
	mu       sync.Mutex
}

func NewDB(saveFile string) DB {
	return DB{saveFile: saveFile, nextUser: 0}
}

func (db *DB) Load() {
	db.mu.Lock()
	defer db.mu.Unlock()
	jsonFile, err := os.Open(db.saveFile)
	if err != nil {
		logger.Error.Println("db load error: ", err)
		return
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		logger.Error.Println("db load error: ", err)
	}
	err = json.Unmarshal(byteValue, &db.users)
	if err != nil {
		logger.Error.Println("db load error: ", err)
	}
	logger.General.Println("db loaded")
}

func (db *DB) save() {
	jsonFile, err := os.Create(db.saveFile)
	if err != nil {
		logger.Error.Println("db save error: ", err)
		return
	}
	defer jsonFile.Close()

	byteValue, err := json.Marshal(db.users)

	if err != nil {
		logger.Error.Println("db save error: ", err)
		return
	}
	jsonFile.Write(byteValue)
	logger.General.Println("db saved")
}

func (db *DB) AddUser(user User) {
	db.mu.Lock()
	db.users = append(db.users, user)
	db.save()
	db.mu.Unlock()
}

func (db *DB) NextUser() *User {
	db.mu.Lock()
	defer db.mu.Unlock()
	if len(db.users) == 0 {
		return nil
	}

	user := db.users[db.nextUser]
	updatedUser := user
	updatedUser.LastCheck = time.Now().Format("2006-01-02 15:04 -0700 MST")
	db.users[db.nextUser] = updatedUser
	defer db.save()
	db.nextUser = (db.nextUser + 1) % len(db.users)
	return &user
}
