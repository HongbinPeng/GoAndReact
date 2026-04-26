package config

import (
	"os"
	"strconv"
)

const defaultMaxUploadSize int64 = 10 * 1024 * 1024

type AppConfig struct {
	Addr          string
	DBPath        string
	UploadDir     string
	JWTSecret     string
	MaxUploadSize int64
}

func Load() AppConfig {
	return AppConfig{
		Addr:          getEnv("APP_ADDR", ":8080"),
		DBPath:        getEnv("DB_PATH", "photo_album.db"),
		UploadDir:     getEnv("UPLOAD_DIR", "uploads"),
		JWTSecret:     getEnv("JWT_SECRET", "penghongbin"),
		MaxUploadSize: getEnvAsInt64("MAX_UPLOAD_SIZE", defaultMaxUploadSize),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func getEnvAsInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}

	return parsed
}
