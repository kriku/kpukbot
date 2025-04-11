package repository

import (
	"github.com/kriku/kpukbot/internal/models"
)

// Repository defines the data persistence operations for the application
type Repository interface {
	// User-related operations
	SaveUser(models.User) error
	GetUser(id int64) (*models.User, error)

	// Thread-related operations
	SaveThread(t models.Thread) error
	GetThread(id int64) (*models.Thread, error)

	// Message-related operations
	SaveMessage(threadID int64, m models.Message) error
	GetMessages(threadID int64) ([]*models.Message, error)

	// Connection management
	Close() error
}
