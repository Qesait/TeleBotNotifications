package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type SpotifyConfig struct {
	ClientId     string
	ClientSecret string
	Scope        string `json:"scope"`
	RedirectUri  string `json:"redirect_uri"`
}

type TelegramConfig struct {
	BotToken string
}

type Config struct {
	configName string
	Port       uint          `json:"port"`
	Spotify    SpotifyConfig `json:"spotify"`
	Telegram   TelegramConfig
}

func (c *Config) readEnv() error {
	c.Spotify.ClientId = os.Getenv("SPOTIFY_CLIENT_ID")
	if c.Spotify.ClientId == "" {
		return fmt.Errorf("spotify id not specified")
	}
	c.Spotify.ClientSecret = os.Getenv("SPOTIFY_CLIENT_SECRET")
	if c.Spotify.ClientSecret == "" {
		return fmt.Errorf("spotify secret not specified")
	}
	c.Telegram.BotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	if c.Telegram.BotToken == "" {
		return fmt.Errorf("telegram bot token not specified")
	}

	c.configName = "config"
	if os.Getenv("CONFIG_NAME") != "" {
		c.configName = os.Getenv("CONFIG_NAME")
	}
	c.configName = "configs/" + c.configName
	if !strings.HasSuffix(c.configName, ".json") {
		c.configName = c.configName + ".json"
	}

	return nil
}

func (c *Config) readJson() error {
	log.Println(c.configName)
	jsonFile, err := os.Open(c.configName)
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

	log.Println("config parsed")
	return config, nil
}
