package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"path/filepath"
)

type SpotifyConfig struct {
	ClientId     string
	ClientSecret string
	Scope        string `json:"scope"`
	RedirectUri  string `json:"redirect_uri"`
}

type TelegramConfig struct {
	BotToken    string
	AdminChatId int
	UpdateDelay uint `json:"update_delay"`
}

type LoggerConfig struct {
	Path             string
	TelegramLogLevel uint `json:"telegram_log_level"`
	FileLogLevel     uint `json:"file_log_level"`
	StdLogLevel      uint `json:"std_log_level"`
}

type Config struct {
	WorkingDirectory string
	Port              uint           `json:"port"`
	Spotify           SpotifyConfig  `json:"spotify"`
	Telegram          TelegramConfig `json:"telegram"`
	Logger            LoggerConfig   `json:"logger"`
}

func (c *Config) readEnv() error {
	c.Telegram.BotToken = strings.TrimRight(os.Getenv("TELEGRAM_BOT_TOKEN"), "\r")
	if c.Telegram.BotToken == "" {
		return fmt.Errorf("failed to load config: telegram bot token not specified")
	}
	var err error
	c.Telegram.AdminChatId, err = strconv.Atoi(strings.TrimRight(os.Getenv("TELEGRAM_ADMIN_CHAT_ID"), "\r"))
	if err != nil {
		c.Telegram.AdminChatId = -1
	}
	c.Spotify.ClientId = strings.TrimRight(os.Getenv("SPOTIFY_CLIENT_ID"), "\r")
	if c.Spotify.ClientId == "" {
		return fmt.Errorf("failed to load config: spotify id not specified")
	}
	c.Spotify.ClientSecret = strings.TrimRight(os.Getenv("SPOTIFY_CLIENT_SECRET"), "\r")
	if c.Spotify.ClientSecret == "" {
		return fmt.Errorf("failed to load config: spotify secret not specified")
	}

	c.WorkingDirectory = strings.TrimRight(os.Getenv("WORKING_DIRECTORY"), "\r")
	if c.WorkingDirectory == "" {
		return fmt.Errorf("failed to load config: working dirrectory not specified")
	}
	c.Logger.Path = fmt.Sprintf("%s/logs", c.WorkingDirectory)

	return nil
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

	err = config.readEnv()
	if err != nil {
		return nil, err
	}

	err = config.readJson()
	if err != nil {
		return nil, err
	}
	return config, nil
}
