package config

import "os"

// Config holds application-wide settings.
type Config struct {
	DBPath string
	Port   string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	cfg := Config{
		DBPath: "data/meetkat.db",
		Port:   "8080",
	}
	if v := os.Getenv("MEETKAT_DB_PATH"); v != "" {
		cfg.DBPath = v
	}
	if v := os.Getenv("MEETKAT_PORT"); v != "" {
		cfg.Port = v
	}
	return cfg
}
