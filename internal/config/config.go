package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                       string
	DatabaseURL                string
	MaxRetries                 int
	InitialBackoff, MaxBackoff time.Duration
	WorkerCount                int
	RequestBodyLimit           int64
	RatePerSecond              int
}

func Load() Config {
	return Config{Port: env("PORT", "8080"), DatabaseURL: os.Getenv("DATABASE_URL"), MaxRetries: envInt("MAX_RETRIES", 3), InitialBackoff: time.Duration(envInt("INITIAL_BACKOFF_SECONDS", 2)) * time.Second, MaxBackoff: time.Duration(envInt("MAX_BACKOFF_SECONDS", 300)) * time.Second, WorkerCount: envInt("WORKER_COUNT", 4), RequestBodyLimit: int64(envInt("REQUEST_BODY_LIMIT", 1048576)), RatePerSecond: envInt("RATE_PER_SECOND", 100)}
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func envInt(k string, d int) int {
	if v, err := strconv.Atoi(os.Getenv(k)); err == nil && v > 0 {
		return v
	}
	return d
}
