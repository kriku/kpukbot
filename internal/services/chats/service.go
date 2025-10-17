package chats

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kriku/kpukbot/internal/models"
	"github.com/kriku/kpukbot/internal/repository/chats"
)

// ChatsService provides business logic for chat and queue management
type ChatsService struct {
	repository chats.ChatsRepository
	logger     *slog.Logger
}

// NewChatsService creates a new chats service
func NewChatsService(repository chats.ChatsRepository, logger *slog.Logger) *ChatsService {
	return &ChatsService{
		repository: repository,
		logger:     logger,
	}
}

// CreateOrUpdateChat creates a new chat or updates an existing one
func (s *ChatsService) CreateOrUpdateChat(ctx context.Context, chatID int64, title, chatType, description string) error {
	chat := models.Chat{
		ID:          chatID,
		Title:       title,
		Type:        chatType,
		Description: description,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Check if chat already exists
	existingChat, err := s.repository.GetChat(ctx, chatID)
	if err == nil && existingChat != nil {
		// Chat exists, preserve existing data and update basic info
		chat.UserIDs = existingChat.UserIDs
		chat.QuestionQueue = existingChat.QuestionQueue
		chat.CreatedAt = existingChat.CreatedAt
	}

	err = s.repository.SaveChat(ctx, chat)
	if err != nil {
		s.logger.Error("Failed to save chat", "chatID", chatID, "error", err)
		return fmt.Errorf("failed to save chat: %w", err)
	}

	s.logger.Info("Chat saved successfully", "chatID", chatID, "title", title)
	return nil
}

// AddUserToChat adds a user to a chat and optionally to the queue
func (s *ChatsService) AddUserToChat(ctx context.Context, chatID int64, userID int64, autoEnqueue bool) error {
	// Add user to chat
	err := s.repository.AddUserToChat(ctx, chatID, userID)
	if err != nil {
		s.logger.Error("Failed to add user to chat", "chatID", chatID, "userID", userID, "error", err)
		return fmt.Errorf("failed to add user to chat: %w", err)
	}

	// Auto-enqueue if enabled
	if autoEnqueue {
		settings, err := s.repository.GetChatSettings(ctx, chatID)
		if err != nil {
			s.logger.Warn("Failed to get chat settings for auto-enqueue", "chatID", chatID, "error", err)
		} else if settings.AutoEnqueueNewUsers && settings.EnableQuestionRounds {
			err = s.repository.AddToQueue(ctx, chatID, userID)
			if err != nil {
				s.logger.Error("Failed to auto-enqueue user", "chatID", chatID, "userID", userID, "error", err)
			} else {
				s.logger.Info("User auto-enqueued", "chatID", chatID, "userID", userID)
			}
		}
	}

	s.logger.Info("User added to chat", "chatID", chatID, "userID", userID)
	return nil
}

// RemoveUserFromChat removes a user from a chat and queue
func (s *ChatsService) RemoveUserFromChat(ctx context.Context, chatID int64, userID int64) error {
	err := s.repository.RemoveUserFromChat(ctx, chatID, userID)
	if err != nil {
		s.logger.Error("Failed to remove user from chat", "chatID", chatID, "userID", userID, "error", err)
		return fmt.Errorf("failed to remove user from chat: %w", err)
	}

	s.logger.Info("User removed from chat", "chatID", chatID, "userID", userID)
	return nil
}

// EnqueueUser adds a user to the question queue
func (s *ChatsService) EnqueueUser(ctx context.Context, chatID int64, userID int64) error {
	// Check if questions are enabled for this chat
	settings, err := s.repository.GetChatSettings(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat settings: %w", err)
	}

	if !settings.EnableQuestionRounds {
		return fmt.Errorf("question rounds are disabled for this chat")
	}

	err = s.repository.AddToQueue(ctx, chatID, userID)
	if err != nil {
		s.logger.Error("Failed to enqueue user", "chatID", chatID, "userID", userID, "error", err)
		return fmt.Errorf("failed to enqueue user: %w", err)
	}

	s.logger.Info("User enqueued", "chatID", chatID, "userID", userID)
	return nil
}

// DequeueUser removes a user from the question queue
func (s *ChatsService) DequeueUser(ctx context.Context, chatID int64, userID int64) error {
	err := s.repository.RemoveFromQueue(ctx, chatID, userID)
	if err != nil {
		s.logger.Error("Failed to dequeue user", "chatID", chatID, "userID", userID, "error", err)
		return fmt.Errorf("failed to dequeue user: %w", err)
	}

	s.logger.Info("User dequeued", "chatID", chatID, "userID", userID)
	return nil
}

// GetNextUserInQueue gets the next user to ask a question
func (s *ChatsService) GetNextUserInQueue(ctx context.Context, chatID int64) (*models.QueueEntry, error) {
	entry, err := s.repository.GetNextInQueue(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get next user in queue: %w", err)
	}

	return entry, nil
}

// MarkQuestionAsked marks a user as currently being asked a question
func (s *ChatsService) MarkQuestionAsked(ctx context.Context, chatID int64, userID int64, questionID string) error {
	entry := models.QueueEntry{
		UserID:     userID,
		Status:     models.QueueStatusAsking,
		QuestionID: questionID,
	}
	now := time.Now()
	entry.AskedAt = &now

	err := s.repository.UpdateQueueEntry(ctx, chatID, entry)
	if err != nil {
		s.logger.Error("Failed to mark question as asked", "chatID", chatID, "userID", userID, "error", err)
		return fmt.Errorf("failed to mark question as asked: %w", err)
	}

	s.logger.Info("Question marked as asked", "chatID", chatID, "userID", userID, "questionID", questionID)
	return nil
}

// MarkQuestionAnswered marks a user's question as answered and completed
func (s *ChatsService) MarkQuestionAnswered(ctx context.Context, chatID int64, userID int64) error {
	// Get current entry to preserve existing data
	chat, err := s.repository.GetChat(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %w", err)
	}

	var entry models.QueueEntry
	found := false
	for _, queueEntry := range chat.QuestionQueue {
		if queueEntry.UserID == userID {
			entry = queueEntry
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("user not found in queue")
	}

	entry.Status = models.QueueStatusCompleted
	now := time.Now()
	entry.AnsweredAt = &now

	err = s.repository.UpdateQueueEntry(ctx, chatID, entry)
	if err != nil {
		s.logger.Error("Failed to mark question as answered", "chatID", chatID, "userID", userID, "error", err)
		return fmt.Errorf("failed to mark question as answered: %w", err)
	}

	s.logger.Info("Question marked as answered", "chatID", chatID, "userID", userID)
	return nil
}

// SkipUser marks a user as skipped for this round
func (s *ChatsService) SkipUser(ctx context.Context, chatID int64, userID int64, reason string) error {
	// Get current entry to preserve existing data
	chat, err := s.repository.GetChat(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %w", err)
	}

	var entry models.QueueEntry
	found := false
	for _, queueEntry := range chat.QuestionQueue {
		if queueEntry.UserID == userID {
			entry = queueEntry
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("user not found in queue")
	}

	entry.Status = models.QueueStatusSkipped

	err = s.repository.UpdateQueueEntry(ctx, chatID, entry)
	if err != nil {
		s.logger.Error("Failed to skip user", "chatID", chatID, "userID", userID, "error", err)
		return fmt.Errorf("failed to skip user: %w", err)
	}

	s.logger.Info("User skipped", "chatID", chatID, "userID", userID, "reason", reason)
	return nil
}

// GetUserQueuePosition returns a user's position in the queue
func (s *ChatsService) GetUserQueuePosition(ctx context.Context, chatID int64, userID int64) (int, error) {
	position, err := s.repository.GetQueuePosition(ctx, chatID, userID)
	if err != nil {
		return -1, fmt.Errorf("failed to get queue position: %w", err)
	}

	return position, nil
}

// ResetQueue clears completed entries and optionally rebuilds the queue
func (s *ChatsService) ResetQueue(ctx context.Context, chatID int64, fullReset bool) error {
	var err error
	if fullReset {
		err = s.repository.ResetQueue(ctx, chatID)
		s.logger.Info("Queue fully reset", "chatID", chatID)
	} else {
		err = s.repository.ClearCompletedQueue(ctx, chatID)
		s.logger.Info("Completed entries cleared from queue", "chatID", chatID)
	}

	if err != nil {
		s.logger.Error("Failed to reset queue", "chatID", chatID, "fullReset", fullReset, "error", err)
		return fmt.Errorf("failed to reset queue: %w", err)
	}

	return nil
}

// GetChatSettings retrieves settings for a chat
func (s *ChatsService) GetChatSettings(ctx context.Context, chatID int64) (*models.ChatSettings, error) {
	settings, err := s.repository.GetChatSettings(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat settings: %w", err)
	}

	return settings, nil
}

// UpdateChatSettings updates settings for a chat
func (s *ChatsService) UpdateChatSettings(ctx context.Context, settings models.ChatSettings) error {
	err := s.repository.SaveChatSettings(ctx, settings)
	if err != nil {
		s.logger.Error("Failed to update chat settings", "chatID", settings.ChatID, "error", err)
		return fmt.Errorf("failed to update chat settings: %w", err)
	}

	s.logger.Info("Chat settings updated", "chatID", settings.ChatID)
	return nil
}

// SetChatActive sets the active status of a chat
func (s *ChatsService) SetChatActive(ctx context.Context, chatID int64, isActive bool) error {
	err := s.repository.SetChatActive(ctx, chatID, isActive)
	if err != nil {
		s.logger.Error("Failed to set chat active status", "chatID", chatID, "isActive", isActive, "error", err)
		return fmt.Errorf("failed to set chat active status: %w", err)
	}

	s.logger.Info("Chat active status updated", "chatID", chatID, "isActive", isActive)
	return nil
}

// GetActiveChats returns all active chats
func (s *ChatsService) GetActiveChats(ctx context.Context) ([]*models.Chat, error) {
	chats, err := s.repository.GetActiveChats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active chats: %w", err)
	}

	return chats, nil
}

// GetUserChats returns all chats that contain a specific user
func (s *ChatsService) GetUserChats(ctx context.Context, userID int64) ([]*models.Chat, error) {
	chats, err := s.repository.GetChatsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user chats: %w", err)
	}

	return chats, nil
}
