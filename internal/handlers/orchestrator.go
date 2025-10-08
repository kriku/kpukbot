package handlers

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	botModels "github.com/go-telegram/bot/models"
	"github.com/kriku/kpukbot/internal/models"
	"github.com/kriku/kpukbot/internal/services/orchestrator"
)

type OrchestratorHandler struct {
	orchestrator *orchestrator.OrchestratorService
	logger       *slog.Logger
}

func NewOrchestratorHandler(
	orchestrator *orchestrator.OrchestratorService,
	logger *slog.Logger,
) bot.HandlerFunc {
	handler := &OrchestratorHandler{
		orchestrator: orchestrator,
		logger:       logger.With("handler", "orchestrator"),
	}

	return handler.Handle
}

func (h *OrchestratorHandler) Handle(ctx context.Context, b *bot.Bot, update *botModels.Update) {
	// Skip if not a message
	if update.Message == nil {
		return
	}

	h.logger.InfoContext(ctx, "Received Telegram update",
		"message_id", update.Message.ID,
		"chat_id", update.Message.Chat.ID,
		"text", update.Message.Text)

	// Convert Telegram message to our model
	message := models.NewMessageFromTelegramUpdate(update)
	if message == nil {
		h.logger.WarnContext(ctx, "Failed to convert message")
		return
	}

	// Process through orchestrator
	if err := h.orchestrator.ProcessMessage(ctx, message); err != nil {
		h.logger.ErrorContext(ctx, "Failed to process message",
			"error", err,
			"message_id", message.ID)

		// Send error response to user
		_, sendErr := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Sorry, I encountered an error processing your message. Please try again.",
		})
		if sendErr != nil {
			h.logger.ErrorContext(ctx, "Failed to send error message", "error", sendErr)
		}
	}
}
