package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	GeminiAPIKey    string
	TelegramToken   string
	GeminiModelName string
}

// NewConfig loads configuration from environment variables
func NewConfig() *Config {
	modelName := os.Getenv("GEMINI_MODEL")
	if modelName == "" {
		modelName = "gemini-1.5-flash" // Default model
	}

	return &Config{
		GeminiAPIKey:    os.Getenv("GEMINI_API_KEY"),
		TelegramToken:   os.Getenv("TELEGRAM_API_TOKEN"),
		GeminiModelName: modelName,
	}
}
