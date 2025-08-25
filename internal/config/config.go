package config

import (
	"os"
	"strconv"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	JWTSecret         string
	SessionSigningKey string
	CookieName        string
	CookieSecure      bool
	DatabasePath      string
	Port              string
}

// LoadFromEnv loads configuration from environment variables with sensible defaults.
func LoadFromEnv() *Config {
	cfg := &Config{}
	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "replace_with_a_strong_random_secret"
	}
	cfg.SessionSigningKey = os.Getenv("SESSION_SIGNING_KEY")
	cfg.CookieName = os.Getenv("COOKIE_NAME")
	if cfg.CookieName == "" {
		cfg.CookieName = "nostr_oicd_session"
	}
	cfg.DatabasePath = os.Getenv("DATABASE_PATH")
	if cfg.DatabasePath == "" {
		cfg.DatabasePath = "./database/dev.sqlite3"
	}
	cfg.Port = os.Getenv("PORT")
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if v := os.Getenv("COOKIE_SECURE"); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			cfg.CookieSecure = b
		}
	}
	return cfg
}
