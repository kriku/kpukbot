package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-telegram/bot/models"
)

// TelegramUpdateHandler defines an interface for handling Telegram updates
type TelegramUpdateHandler interface {
	HandleUpdate(ctx context.Context, update *models.Update)
}

// WebhookHandler is responsible for handling HTTP webhook requests
type WebhookHandler struct {
	telegramHandler TelegramUpdateHandler
}

// NewWebhookHandler creates a new WebhookHandler
func NewWebhookHandler(telegramHandler TelegramUpdateHandler) *WebhookHandler {
	return &WebhookHandler{
		telegramHandler: telegramHandler,
	}
}

// HandleTelegramWebhook processes incoming webhook requests
func (h *WebhookHandler) HandleTelegramWebhook(res http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	update := models.Update{}

	err := json.NewDecoder(req.Body).Decode(&update)
	if err != nil {
		log.Printf("Error parsing incoming webhook update: %s", err)

		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Bad Request"))

		return
	}

	log.Printf("Received update: %v", update)

	h.telegramHandler.HandleUpdate(ctx, &update)

	res.WriteHeader(http.StatusOK)
	res.Write([]byte("ok"))
}
