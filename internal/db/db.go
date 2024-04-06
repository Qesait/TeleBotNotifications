package db

import (
	"TeleBotNotifications/internal/logger"
	"TeleBotNotifications/internal/spotify"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

type DB struct {
	UserId    int                 `json:"user_id"`
	ChatId    int                 `json:"chat_id"`
	Token     spotify.OAuth2Token `json:"token"`
	LastCheck string              `json:"last_check"`
	saveFile  string
	mu        sync.Mutex
}

func NewDB(saveFile string) DB {
	return DB{saveFile: saveFile}
}

func (db *DB) Load() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	jsonFile, err := os.Open(db.saveFile)
	if err != nil {
		return fmt.Errorf("db load | can't open save file: %w", err)
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return fmt.Errorf("db load | can't read save file: %w", err)
	}
	err = json.Unmarshal(byteValue, &db)
	if err != nil {
		return fmt.Errorf("db load | wrong save file format: %w", err)
	}
	logger.General.Println("db loaded")
	return nil
}

func (db *DB) save() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	jsonFile, err := os.Create(db.saveFile)
	if err != nil {
		return fmt.Errorf("db save | can't open save file: %w", err)
	}
	defer jsonFile.Close()

	byteValue, err := json.MarshalIndent(db, "", "    ")
	if err != nil {
		return fmt.Errorf("db save | can't marshal save data: %w", err)
	}
	_, err = jsonFile.Write(byteValue)
	if err != nil {
		return fmt.Errorf("db save | can't write into save file: %w", err)
	}
	logger.General.Println("db saved")
	return nil
}
