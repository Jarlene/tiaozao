package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	MinIO    MinIOConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type DatabaseConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type MinIOConfig struct {
	Endpoint       string
	PublicEndpoint string
	AccessKey      string
	SecretKey      string
	UseSSL         bool
	Bucket         string
}

func Load() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("SERVER_MODE", "debug"),
		},
		Database: DatabaseConfig{
			DSN: getEnv("DATABASE_DSN", "host=localhost user=flea password=flea_pass dbname=flea_market port=5432 sslmode=disable TimeZone=Asia/Shanghai"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "flea-market-jwt-secret-key-change-in-production-32chars"),
			AccessTokenTTL:  getEnvDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
			RefreshTokenTTL: getEnvDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour),
		},
		MinIO: MinIOConfig{
			Endpoint:       getEnv("MINIO_ENDPOINT", "localhost:9000"),
			PublicEndpoint: getEnv("MINIO_PUBLIC_ENDPOINT", ""),
			AccessKey:      getEnv("MINIO_ACCESS_KEY", "flea_admin"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "flea_admin_pass"),
			UseSSL:    getEnvBool("MINIO_USE_SSL", false),
			Bucket:    getEnv("MINIO_BUCKET", "flea-avatars"),
		},
	}

	cfg.warnDefaults()
	return cfg
}

// warnDefaults 在启动时检查关键配置使用默认值的情况
func (c *Config) warnDefaults() {
	defaultSecret := "flea-market-jwt-secret-key-change-in-production-32chars"
	if c.JWT.Secret == defaultSecret {
		log.Println("[WARN] JWT_SECRET is using default value, please set it via environment variable in production")
	}
	if c.MinIO.AccessKey == "flea_admin" {
		log.Println("[WARN] MINIO_ACCESS_KEY is using default value, please set it via environment variable in production")
	}
	if c.MinIO.SecretKey == "flea_admin_pass" {
		log.Println("[WARN] MINIO_SECRET_KEY is using default value, please set it via environment variable in production")
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}
