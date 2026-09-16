package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	ServerAddr  string
	DatabaseURL string
	JWTSecret   string
	JWTTTL      time.Duration
}

func Load() (Config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return Config{}, errors.New("JWT_SECRET is not set")
	}

	jwtTTL := os.Getenv("JWT_TTL")
	if jwtTTL == "" {
		return Config{}, errors.New("JWT_TTL is not set")
	}

	ttl, err := time.ParseDuration(jwtTTL)
	if err != nil {
		return Config{}, err
	}

	return Config{
		ServerAddr:  ":8081",
		DatabaseURL: databaseURL,
		JWTSecret:   jwtSecret,
		JWTTTL:      ttl,
	}, nil
}
