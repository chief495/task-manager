package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

// AppConfig — структура конфигурации приложения
type AppConfig struct {
	DBDriver   string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	ServerPort string
	JWTKey     string
}

// LoadConfig загружает конфигурацию из .env или переменных окружения
func LoadConfig() (*AppConfig, error) {
	// Загружаем .env (если есть)
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	return &AppConfig{
		DBDriver:   getEnv("DB_DRIVER", "postgres"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "task_manager"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
		JWTKey:     getEnv("JWT_KEY", "secret-key"),
	}, nil
}

// getEnv — вспомогательная функция для получения переменной окружения
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
