package kpukbot

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-telegram/bot/models"
	"github.com/kriku/kpukbot/app"
	internalModels "github.com/kriku/kpukbot/internal/models"
)

func HandleTelegramWebhook(res http.ResponseWriter, req *http.Request) {
	// Use request context with the Cloud Run function lifecycle
	ctx := req.Context()

	a, err := app.InitApp(ctx)
	if err != nil {
		log.Printf("Failed to initialize application: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("initialization error"))
		return
	}
	defer a.Close()

	// Decode the webhook update directly
	var update models.Update
	if err := json.NewDecoder(req.Body).Decode(&update); err != nil {
		log.Printf("Failed to decode webhook update: %v", err)
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("invalid request body"))
		return
	}

	// Respond to Telegram immediately to avoid timeout
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("ok"))

	// Process the message synchronously within the function's lifetime
	if update.Message != nil {
		message := internalModels.NewMessageFromTelegramUpdate(&update)
		if message != nil {
			if err := a.Orchestrator.ProcessMessage(ctx, message); err != nil {
				log.Printf("Failed to process message: %v", err)
			}
		}
	}
}
