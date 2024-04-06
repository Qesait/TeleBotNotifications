package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	// "strconv"
	// "strings"
)

type SpotifyConfig struct {
	ClientId     string `env:"SPOTIFY_CLIENT_ID"`
	ClientSecret string `env:"SPOTIFY_CLIENT_SECRET"`
	Scope        string `json:"scope"`
	RedirectUri  string `json:"redirect_uri"`
}

type TelegramConfig struct {
	BotToken string `env:"TELEGRAM_BOT_TOKEN"`
	ChatId   int    `env:"TELEGRAM_CHAT_ID"`
	Timeout  int    `json:"timeout"`
}

type LoggerConfig struct {
	Path             string
	TelegramLogLevel uint `json:"telegram_log_level"`
	FileLogLevel     uint `json:"file_log_level"`
	StdLogLevel      uint `json:"std_log_level"`
}

type Config struct {
	WorkingDirectory string         `env:"WORKING_DIRECTORY"`
	Port             uint           `json:"port"`
	Spotify          SpotifyConfig  `json:"spotify"`
	Telegram         TelegramConfig `json:"telegram"`
	Logger           LoggerConfig   `json:"logger"`
}

func (c *Config) readJson() error {
	jsonFile, err := os.Open(filepath.Join(c.WorkingDirectory, "configs", "config.json"))
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteValue, c)
	if err != nil {
		return err
	}

	return nil
}

func NewConfig() (*Config, error) {
	var err error
	config := &Config{}

	if err := LoadConfigFromEnv(config); err != nil {
		return nil, fmt.Errorf("error loading env: %v", err)
	}

	err = config.readJson()
	if err != nil {
		return nil, fmt.Errorf("error loading config: %v", err)
	}
	return config, nil
}
