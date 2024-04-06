package db

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"TeleBotNotifications/internal/logger"
	"TeleBotNotifications/internal/spotify"
)

type User struct {
	UserId    int                 `json:"user_id"`
	ChatId    int                 `json:"chat_id"`
	Token     spotify.OAuth2Token `json:"token"`
	LastCheck time.Time           `json:"last_check"`
}

type DB struct {
	user     *User
	saveFile string
	mu       sync.Mutex
}

func NewDB(saveFile string) DB {
	return DB{saveFile: saveFile}
}

func (db *DB) Load() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	jsonFile, err := os.Open(db.saveFile)
	if err != nil {
		return fmt.Errorf("can't open save file: %w", err)
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return fmt.Errorf("can't read save file: %w", err)
	}
	user := &User{}
	err = json.Unmarshal(byteValue, user)
	if err != nil {
		return fmt.Errorf("wrong save file format: %w", err)
	}
	db.user = user
	logger.General.Println("db loaded")
	return nil
}

func (db *DB) Save() error {
	if db.user == nil {
		return nil
	}
	db.mu.Lock()
	defer db.mu.Unlock()

	jsonFile, err := os.Create(db.saveFile)
	if err != nil {
		return fmt.Errorf("can't open save file: %w", err)
	}
	defer jsonFile.Close()

	byteValue, err := json.MarshalIndent(db.user, "", "    ")
	if err != nil {
		return fmt.Errorf("can't marshal save data: %w", err)
	}
	_, err = jsonFile.Write(byteValue)
	if err != nil {
		return fmt.Errorf("can't write into save file: %w", err)
	}
	logger.General.Println("db saved")
	return nil
}

func (db *DB) Set(newUser User) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.user = &newUser
}

func (db *DB) Get() *User {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.user == nil {
		return nil
	}
	userCopy := *db.user
	return &userCopy
}
