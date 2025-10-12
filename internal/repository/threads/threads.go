package threads

import (
	"context"

	"github.com/kriku/kpukbot/internal/models"
)

type ThreadsRepository interface {
	SaveThread(ctx context.Context, thread *models.Thread) error
	GetThread(ctx context.Context, id string) (*models.Thread, error)
	GetThreadByMessageID(ctx context.Context, messageID int) (*models.Thread, error)
	GetThreadsByChatID(ctx context.Context, chatID int64) ([]*models.Thread, error)
	GetActiveThreadsByChatID(ctx context.Context, chatID int64) ([]*models.Thread, error)
	UpdateThread(ctx context.Context, thread *models.Thread) error
	Close() error
}
