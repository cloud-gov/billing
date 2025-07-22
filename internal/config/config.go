package config

import (
	"errors"
	"os"
)

type Config struct {
	ApiUrl         string
	CFClientId     string
	CFClientSecret string
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

	return c, nil
}
