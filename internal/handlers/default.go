package handlers

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/kriku/kpukbot/internal/repository/messages"
)

type DefaultHandler bot.HandlerFunc

func NewDefaultHandler(logger *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		logger.Info("Received update: %v", update)

		messages.MessagesRepository.SaveMessage(ctx, update)

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Sorry, I didn't understand that command.",
		})
	}
}
