package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBDriver   string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSchema   string
	Port          string
	ReposDir      string
	AdminPassword string
	JWTSecret     string
	BasePath      string
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	basePath := getEnv("BASE_PATH", "")
	basePath = strings.TrimRight(basePath, "/")
	if basePath != "" && !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}

	return &Config{
		DBDriver:   getEnv("DB_DRIVER", "sqlite"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "1433"),
		DBUser:     getEnv("DB_USER", "sa"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "deployment_logs"),
		DBSchema:   getEnv("DB_SCHEMA", "deploymentlogs"),
		Port:          getEnv("PORT", "3000"),
		ReposDir:      getEnv("REPOS_DIR", "./repos"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin"),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		BasePath:      basePath,
	}, nil
}
