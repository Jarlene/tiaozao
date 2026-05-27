package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ServerPort   string
	DatabaseURL  string
	JWTSecret    string
	ImageDir     string
	ImageBaseURL string
}

func Load() *Config {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "tiaozao")
	password := getEnv("DB_PASSWORD", "tiaozao123")
	dbname := getEnv("DB_NAME", "tiaozao")

	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&TimeZone=Asia/Shanghai",
		user, password, host, port, dbname)

	return &Config{
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", databaseURL),
		JWTSecret:    getEnv("JWT_SECRET", "tiaozao-dev-secret"),
		ImageDir:     getEnv("IMAGE_DIR", "./uploads"),
		ImageBaseURL: getEnv("IMAGE_BASE_URL", "http://localhost:8080/uploads"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
