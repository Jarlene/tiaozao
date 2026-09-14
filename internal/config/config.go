package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	ServerPort   string
	DatabaseURL  string
	JWTSecret    string
	ImageDir     string
	ImageBaseURL string
	// CORSAllowedOrigins lists the origins allowed to call the API in
	// production. Comma-separated; defaults to the local frontend.
	CORSAllowedOrigins []string
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
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", databaseURL),
		JWTSecret:          getEnv("JWT_SECRET", "tiaozao-dev-secret"),
		ImageDir:           getEnv("IMAGE_DIR", "./uploads"),
		ImageBaseURL:       getEnv("IMAGE_BASE_URL", "http://localhost:8080/uploads"),
		CORSAllowedOrigins: getEnvList("CORS_ALLOWED_ORIGINS", "http://localhost:8080"),
	}
}

// getEnv reads an environment variable or returns the fallback default.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getEnvList parses a comma-separated environment variable into a slice.
func getEnvList(key, fallback string) []string {
	v := os.Getenv(key)
	if v == "" {
		v = fallback
	}
	var out []string
	for _, item := range strings.Split(v, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}
