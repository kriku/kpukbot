package handlers

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	botModels "github.com/go-telegram/bot/models"

	"github.com/kriku/kpukbot/internal/models"
	messages "github.com/kriku/kpukbot/internal/repository/messages"
)

type DefaultHandler struct {
	l *slog.Logger
	m messages.MessagesRepository
}

func NewDefaultHandler(l *slog.Logger, m messages.MessagesRepository) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *botModels.Update) {
		l.Info("Received update: %v", update)

		message := models.NewMessageFromTelegramUpdate(update)

		messages, err := m.GetMessages(ctx, message.ChatID)

		if err != nil {
			l.Error("Failed to get messages from DB: %v", err)
			return
		}

		l.Info("Total messages in DB: %d", len(messages))
		m.SaveMessage(ctx, *message)

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Sorry, I didn't understand that command.",
		})
	}
}
