package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	DatabaseURL  string
	IdleTTL      time.Duration
	SessionTTL   time.Duration
	CookieSecure bool
}

func Load() (Config, error) {
	_ = godotenv.Load()

	idleTTL, err := envDuration("SESSION_IDLE_TTL", 30*time.Minute)
	if err != nil {
		return Config{}, err
	}
	sessionTTL, err := envDuration("SESSION_TTL", 12*time.Hour)
	if err != nil {
		return Config{}, err
	}
	if idleTTL >= sessionTTL {
		return Config{}, fmt.Errorf(
			"config: SESSION_IDLE_TTL (%s) must be shorter than SESSION_TTL (%s)",
			idleTTL,
			sessionTTL,
		)
	}
	cookieSecure, err := envBool("COOKIE_SECURE", false)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Port:         os.Getenv("PORT"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		IdleTTL:      idleTTL,
		SessionTTL:   sessionTTL,
		CookieSecure: cookieSecure,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is not set")
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg, nil
}

func envDuration(key string, def time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return def, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf(
			"config: %s must be a duration like \"30m\" or \"1h30m\": %w",
			key,
			err,
		)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("config: %s can't be 0 or a negative value", key)
	}
	return duration, nil
}

func envBool(key string, def bool) (bool, error) {
	value := os.Getenv(key)
	if value == "" {
		return def, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("config: %s must be a valid boolean: true / false", key)
	}
	return parsed, nil
}
