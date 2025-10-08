package messages

import (
	"context"
	"log/slog"

	tmodels "github.com/go-telegram/bot/models"

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

func (s *TelegramMessagesService) HandleMessage(ctx context.Context, update *tmodels.Update) error {
	message := models.NewMessageFromTelegramUpdate(update)

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
