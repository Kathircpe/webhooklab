package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL        string
	HTTPAddr           string
	BaseURL            string
	DefaultTTL         time.Duration
	MaxBodyBytes       int64
	CleanupInterval    time.Duration
	ExpiredRetention   time.Duration
	TrustedProxies     string
	CORSOrigins        string
	EventCap           int
	ReplayTimeout        time.Duration
	ReplayMaxResponse    int64
	ReplayAllowPrivate   bool
}

func Load() Config {
	c := Config{
		DatabaseURL:      getenv("DATABASE_URL", "postgres://webhooklab:webhooklab@localhost:5432/webhooklab?sslmode=disable"),
		HTTPAddr:         getenv("HTTP_ADDR", ":8080"),
		BaseURL:          getenv("WEBHOOKLAB_BASE_URL", "http://localhost:8080"),
		DefaultTTL:       mustDuration(getenv("WEBHOOKLAB_DEFAULT_TTL", "168h")),
		MaxBodyBytes:     mustInt64(getenv("WEBHOOKLAB_MAX_BODY_BYTES", "1048576")),
		CleanupInterval:  mustDuration(getenv("WEBHOOKLAB_CLEANUP_INTERVAL", "1h")),
		ExpiredRetention: mustDuration(getenv("WEBHOOKLAB_EXPIRED_RETENTION", "720h")),
		TrustedProxies:   getenv("WEBHOOKLAB_TRUSTED_PROXIES", ""),
		CORSOrigins:      getenv("WEBHOOKLAB_CORS_ORIGINS", "*"),
		EventCap:         mustInt(getenv("WEBHOOKLAB_EVENT_CAP", "1000")),
		ReplayTimeout:      mustDuration(getenv("WEBHOOKLAB_REPLAY_TIMEOUT", "10s")),
		ReplayMaxResponse:  mustInt64(getenv("WEBHOOKLAB_REPLAY_MAX_RESPONSE_BYTES", "1048576")),
		ReplayAllowPrivate: getenv("WEBHOOKLAB_REPLAY_ALLOW_PRIVATE", "false") == "true",
	}
	return c
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func mustDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return d
}

func mustInt64(s string) int64 {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return n
}

func mustInt(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return n
}
