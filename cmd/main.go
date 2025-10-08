package main

import (
	"log"

	"github.com/kriku/kpukbot/app"
)

// Send any text message to the bot after the bot has been started
func main() {
	a, err := app.InitAppLocal()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer a.Close()

	a.Logger.Info("Starting Telegram bot with Gemini integration...")
	a.Logger.Info("Orchestrator initialized successfully")

	a.MessengerClient.Start()
}
