package chats

import (
	"context"

	"github.com/kriku/kpukbot/internal/models"
)

type ChatsRepository interface {
	// SaveChat saves or updates a chat in the repository
	SaveChat(ctx context.Context, chat models.Chat) error

	// GetChat retrieves a chat by its ID
	GetChat(ctx context.Context, chatID int64) (*models.Chat, error)

	// GetAllChats retrieves all chats
	GetAllChats(ctx context.Context) ([]*models.Chat, error)

	// GetActiveChats retrieves all active chats
	GetActiveChats(ctx context.Context) ([]*models.Chat, error)

	// UpdateChatUsers updates the list of users in a chat
	UpdateChatUsers(ctx context.Context, chatID int64, userIDs []int64) error

	// AddUserToChat adds a user to the chat's user list
	AddUserToChat(ctx context.Context, chatID int64, userID int64) error

	// RemoveUserFromChat removes a user from the chat's user list
	RemoveUserFromChat(ctx context.Context, chatID int64, userID int64) error

	// Queue Management
	// UpdateQuestionQueue updates the entire question queue for a chat
	UpdateQuestionQueue(ctx context.Context, chatID int64, queue []models.QueueEntry) error

	// AddToQueue adds a user to the question queue
	AddToQueue(ctx context.Context, chatID int64, userID int64) error

	// RemoveFromQueue removes a user from the question queue
	RemoveFromQueue(ctx context.Context, chatID int64, userID int64) error

	// GetNextInQueue gets the next user in the question queue
	GetNextInQueue(ctx context.Context, chatID int64) (*models.QueueEntry, error)

	// UpdateQueueEntry updates a specific queue entry
	UpdateQueueEntry(ctx context.Context, chatID int64, entry models.QueueEntry) error

	// GetQueuePosition gets a user's current position in the queue
	GetQueuePosition(ctx context.Context, chatID int64, userID int64) (int, error)

	// ClearCompletedQueue removes completed entries from the queue
	ClearCompletedQueue(ctx context.Context, chatID int64) error

	// ResetQueue clears the entire queue and rebuilds it from active users
	ResetQueue(ctx context.Context, chatID int64) error

	// Chat Settings
	// SaveChatSettings saves chat settings
	SaveChatSettings(ctx context.Context, settings models.ChatSettings) error

	// GetChatSettings retrieves chat settings
	GetChatSettings(ctx context.Context, chatID int64) (*models.ChatSettings, error)

	// Chat Status
	// SetChatActive sets the active status of a chat
	SetChatActive(ctx context.Context, chatID int64, isActive bool) error

	// GetChatsByUser retrieves all chats that contain a specific user
	GetChatsByUser(ctx context.Context, userID int64) ([]*models.Chat, error)
}
