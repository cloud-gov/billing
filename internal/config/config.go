package config

import (
	"errors"
	"log/slog"
	"os"
)

type Config struct {
	CFApiUrl string
	// CFClientID is the ID of the client created in UAA with permission to read data from CAPI. Because the billing service is the resource server in the OIDC relationship, it doubles as the audience claim in JWTs.
	CFClientId     string
	CFClientSecret string
	Host           string
	Port           string
	LogLevel       slog.Level
	Issuer         string
}

func New() (Config, error) {
	c := Config{}
	c.CFApiUrl = os.Getenv("CF_API_URL")
	if c.CFApiUrl == "" {
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
	c.Issuer = os.Getenv("OIDC_ISSUER")
	if c.Issuer == "" {
		return Config{}, errors.New("reading OIDC_ISSUER")
	}
	return c, nil
}
