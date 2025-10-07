package messages

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-telegram/bot"
	telegramModels "github.com/go-telegram/bot/models"

	"github.com/kriku/kpukbot/internal/models"
	repositories "github.com/kriku/kpukbot/internal/repository/messages"
)

type TelegramMessagesService struct {
	repo   repositories.MessagesRepository
	logger *slog.Logger
}

func NewTelegramMessagesService(repo repositories.MessagesRepository, logger *slog.Logger) *TelegramMessagesService {
	return &TelegramMessagesService{
		logger: logger.With("service", "messages"),
		repo:   repo,
	}
}

func (s *TelegramMessagesService) HandleUpdate(ctx context.Context, update *telegramModels.Update) error {
	if update.Message == nil {
		return nil
	}

	msg := update.Message
	message := &models.Message{
		ID:     msg.ID,
		ChatID: msg.Chat.ID,
		Date:   time.Unix(int64(msg.Date), 0),
		Text:   msg.Text,
	}

	if msg.From != nil {
		message.UserID = msg.From.ID
		message.Username = msg.From.Username
		message.FirstName = msg.From.FirstName
		message.LastName = msg.From.LastName
		message.IsBot = msg.From.IsBot
	}

	s.logger.InfoContext(ctx, "Processing message",
		"message_id", message.ID,
		"chat_id", message.ChatID,
		"user_id", message.UserID,
		"text_length", len(message.Text),
	)

	err := s.repo.SaveMessage(ctx, *message)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to save message",
			"error", err,
			"message_id", message.ID,
		)
		return err
	}

	s.logger.DebugContext(ctx, "Message saved successfully", "message_id", message.ID)
	return nil
}

func (s *TelegramMessagesService) DefaultHandler(ctx context.Context, bot *bot.Bot, update *telegramModels.Update) {
	s.HandleUpdate(ctx, update)
}
