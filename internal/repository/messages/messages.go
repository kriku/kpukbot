package messages

import (
	"context"

	"github.com/kriku/kpukbot/internal/models"
)

type MessagesRepository interface {
	SaveMessage(ctx context.Context, m models.Message) error
	GetMessage(ctx context.Context, ID int64) ([]*models.Message, error)
	GetMessages(ctx context.Context, chatID int64) ([]*models.Message, error)
}
