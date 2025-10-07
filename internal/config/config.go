package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	GeminiAPIKey    string
	TelegramToken   string
	GeminiModelName string
	FilestoreConfig FirestoreConfig
}

// FirestoreConfig holds the configuration for Firebase/Firestore
type FirestoreConfig struct {
	ProjectID    string
	EmulatorHost string // For local development
	UseEmulator  bool
}

// NewConfig loads configuration from environment variables
func NewConfig() *Config {
	modelName := os.Getenv("GEMINI_MODEL")
	if modelName == "" {
		modelName = "gemini-2.5-flash" // Default model
	}

	// Load Firestore configuration
	firestoreConfig := FirestoreConfig{
		ProjectID:    os.Getenv("CLOUD_PROJECT_ID"),
		EmulatorHost: os.Getenv("FIRESTORE_EMULATOR_HOST"),
		UseEmulator:  os.Getenv("USE_FIRESTORE_EMULATOR") == "true",
	}

	return &Config{
		GeminiAPIKey:    os.Getenv("GEMINI_API_KEY"),
		TelegramToken:   os.Getenv("TELEGRAM_API_TOKEN"),
		GeminiModelName: modelName,
		FilestoreConfig: firestoreConfig,
	}
}
