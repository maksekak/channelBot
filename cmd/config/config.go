package config

import (
	"os"
)

type Config struct {
	TelegramBotToken string
	AdminUsername    string
	AdminPassword    string
	WebPort          string
	DatabaseURL      string
	Timezone         string
	JWTSecret        string
	LogLevel         string
}

func Load() *Config {
	return &Config{
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		AdminUsername:    getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:    getEnv("ADMIN_PASSWORD", "admin"),
		WebPort:          getEnv("WEB_PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://user:pass@localhost/mgok_bot"),
		Timezone:         getEnv("TIMEZONE", "Europe/Moscow"),
		JWTSecret:        getEnv("JWT_SECRET", "your-secret-key"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
