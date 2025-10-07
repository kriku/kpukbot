package handlers

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	messages "github.com/kriku/kpukbot/internal/repository/messages"
)

type DefaultHandler struct {
	l *slog.Logger
	m messages.MessagesRepository
}

func NewDefaultHandler(l *slog.Logger, m messages.MessagesRepository) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		l.Info("Received update: %v", update)

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Sorry, I didn't understand that command.",
		})
	}
}
