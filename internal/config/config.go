package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppName string
	AppEnv  string
	AppPort string

	DatabaseURL string

	HTTPReadTimeout     time.Duration
	HTTPWriteTimeout    time.Duration
	HTTPIdleTimeout     time.Duration
	HTTPShutdownTimeout time.Duration

	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnMaxIdleTime time.Duration

	RedditBaseURL        string
	RedditUserAgent      string
	RedditRequestTimeout time.Duration
	RedditDefaultLimit   int
	RedditMaxLimit       int

	HNBaseURL        string
	HNRequestTimeout time.Duration
	HNDefaultLimit   int
	HNMaxLimit       int
}

func Load() (Config, error) {
	cfg := Config{
		AppName:     getEnv("APP_NAME", "synaptica-api"),
		AppEnv:      getEnv("APP_ENV", "development"),
		AppPort:     getEnvWithFallbacks("8080", "APP_PORT", "PORT"),
		DatabaseURL: os.Getenv("DATABASE_URL"),

		RedditBaseURL:   getEnv("REDDIT_BASE_URL", ""),
		RedditUserAgent: getEnv("REDDIT_USER_AGENT", ""),
		HNBaseURL:       getEnv("HN_BASE_URL", "https://hacker-news.firebaseio.com"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	var err error
	if cfg.HTTPReadTimeout, err = parseDurationEnv("HTTP_READ_TIMEOUT", "10s"); err != nil {
		return Config{}, err
	}
	if cfg.HTTPWriteTimeout, err = parseDurationEnv("HTTP_WRITE_TIMEOUT", "120s"); err != nil {
		return Config{}, err
	}
	if cfg.HTTPIdleTimeout, err = parseDurationEnv("HTTP_IDLE_TIMEOUT", "60s"); err != nil {
		return Config{}, err
	}
	if cfg.HTTPShutdownTimeout, err = parseDurationEnv("HTTP_SHUTDOWN_TIMEOUT", "10s"); err != nil {
		return Config{}, err
	}
	if cfg.DBMaxOpenConns, err = parseIntEnv("DB_MAX_OPEN_CONNS", "10"); err != nil {
		return Config{}, err
	}
	if cfg.DBMaxIdleConns, err = parseIntEnv("DB_MAX_IDLE_CONNS", "5"); err != nil {
		return Config{}, err
	}
	if cfg.DBConnMaxLifetime, err = parseDurationEnv("DB_CONN_MAX_LIFETIME", "30m"); err != nil {
		return Config{}, err
	}
	if cfg.DBConnMaxIdleTime, err = parseDurationEnv("DB_CONN_MAX_IDLE_TIME", "10m"); err != nil {
		return Config{}, err
	}
	if cfg.RedditRequestTimeout, err = parseDurationEnv("REDDIT_REQUEST_TIMEOUT", "15s"); err != nil {
		return Config{}, err
	}
	if cfg.RedditDefaultLimit, err = parsePositiveIntEnv("REDDIT_DEFAULT_LIMIT", "25"); err != nil {
		return Config{}, err
	}
	if cfg.RedditMaxLimit, err = parsePositiveIntEnv("REDDIT_MAX_LIMIT", "100"); err != nil {
		return Config{}, err
	}
	if cfg.HNRequestTimeout, err = parseDurationEnv("HN_REQUEST_TIMEOUT", "20s"); err != nil {
		return Config{}, err
	}
	if cfg.HNDefaultLimit, err = parsePositiveIntEnv("HN_DEFAULT_LIMIT", "25"); err != nil {
		return Config{}, err
	}
	if cfg.HNMaxLimit, err = parsePositiveIntEnv("HN_MAX_LIMIT", "100"); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvWithFallbacks(fallback string, keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return fallback
}

func parseDurationEnv(key, fallback string) (time.Duration, error) {
	value := getEnv(key, fallback)
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}
	return duration, nil
}

func parseIntEnv(key, fallback string) (int, error) {
	value := getEnv(key, fallback)
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer: %w", key, err)
	}
	if parsed < 0 {
		return 0, fmt.Errorf("%s must be greater than or equal to 0", key)
	}
	return parsed, nil
}

func parsePositiveIntEnv(key, fallback string) (int, error) {
	parsed, err := parseIntEnv(key, fallback)
	if err != nil {
		return 0, err
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be greater than 0", key)
	}
	return parsed, nil
}
