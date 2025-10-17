package messages

import (
	"context"
	"log/slog"
	"time"

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

func (s *TelegramMessagesService) SaveBotMessage(ctx context.Context, chatID int64, messageID int, text string) error {
	message := &models.Message{
		ID:        messageID,
		ChatID:    chatID,
		UserID:    0, // Bot has no user ID
		Text:      text,
		Username:  "kpukbot",
		FirstName: "",
		LastName:  "",
		Date:      time.Now(),
		IsBot:     true,
	}

	s.logger.InfoContext(ctx, "Saving bot message",
		"message_id", message.ID,
		"chat_id", message.ChatID,
		"text_length", len(message.Text),
	)

	err := s.repo.SaveMessage(ctx, *message)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to save bot message",
			"error", err,
			"message_id", message.ID,
		)
		return err
	}

	s.logger.DebugContext(ctx, "Bot message saved successfully", "message_id", message.ID)
	return nil
}
