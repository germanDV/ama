package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Domain    string
	Port      int
	Secure    bool
	TTL       time.Duration
	RedisHost string
	RedisPort int
	RedisPass string
	LogLevel  slog.Level
}

func loadConfig() (*AppConfig, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		return nil, errors.New("env var DOMAIN not set")
	}

	portStr := os.Getenv("PORT")
	if portStr == "" {
		return nil, errors.New("env var PORT not set")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	secureStr := os.Getenv("SECURE")
	if secureStr == "" {
		return nil, errors.New("env var SECURE not set")
	}
	secure, err := strconv.ParseBool(secureStr)
	if err != nil {
		return nil, err
	}

	ttlStr := os.Getenv("TTL")
	if ttlStr == "" {
		return nil, errors.New("env var TTL not set")
	}
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		return nil, err
	}

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		return nil, errors.New("env var REDIS_HOST not set")
	}

	redisPortStr := os.Getenv("REDIS_PORT")
	if redisPortStr == "" {
		return nil, errors.New("env var REDIS_PORT not set")
	}
	redisPort, err := strconv.Atoi(redisPortStr)
	if err != nil {
		return nil, err
	}

	redisPass := os.Getenv("REDIS_PASS")

	logLevelStr := os.Getenv("LOG_LEVEL")
	if logLevelStr == "" {
		return nil, errors.New("env var LOG_LEVEL not set")
	}
	logLevel, err := levelFromStr(logLevelStr)
	if err != nil {
		return nil, err
	}

	return &AppConfig{
		Domain:    domain,
		Port:      port,
		Secure:    secure,
		TTL:       ttl,
		RedisHost: redisHost,
		RedisPort: redisPort,
		RedisPass: redisPass,
		LogLevel:  logLevel,
	}, nil
}

func levelFromStr(str string) (slog.Level, error) {
	switch str {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelDebug, fmt.Errorf("invalid log level: %s", str)
	}
}
