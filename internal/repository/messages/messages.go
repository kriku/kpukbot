package messages

import (
	"context"

	"github.com/kriku/kpukbot/internal/models"
)

type MessagesRepository interface {
	SaveMessage(ctx context.Context, m models.Message) error
	GetMessage(ctx context.Context, id int64) ([]*models.Message, error)

	// GetMessagesByChatID(ctx context.Context, chatID int64, limit int) ([]*models.Message, error)

	Close() error
}
