package config

import (
	"errors"
	"log/slog"
	"os"
)

type Config struct {
	ApiUrl         string
	CFClientId     string
	CFClientSecret string
	Host           string
	Port           string
	LogLevel       slog.Level
}

func New() (Config, error) {
	c := Config{}
	c.ApiUrl = os.Getenv("CF_API_URL")
	if c.ApiUrl == "" {
		return Config{}, errors.New("reading CF_API_URL")
	}
	c.CFClientId = os.Getenv("CF_CLIENT_ID")
	if c.CFClientId == "" {
		return Config{}, errors.New("reading CF_CLIENT_ID")
	}
	c.CFClientSecret = os.Getenv("CF_CLIENT_SECRET")
	if c.CFClientSecret == "" {
		return Config{}, errors.New("reading CF_CLIENT_SECRET")
	}
	c.Port = os.Getenv("PORT")
	if c.Port == "" {
		c.Port = "8080"
	}
	c.Host = os.Getenv("HOST")

	levelString := os.Getenv("LOG_LEVEL")
	err := c.LogLevel.UnmarshalText([]byte(levelString))
	if err != nil {
		c.LogLevel = slog.LevelInfo
	}
	return c, nil
}
