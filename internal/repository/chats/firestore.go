package chats

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/kriku/kpukbot/internal/models"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	chatsCollection    = "chats"
	settingsCollection = "chat_settings"
)

// FirestoreRepository implements ChatsRepository interface using Firestore
type FirestoreRepository struct {
	client *firestore.Client
}

// NewFirestoreChatsRepository creates a new FirestoreRepository with existing client
func NewFirestoreChatsRepository(client *firestore.Client) ChatsRepository {
	return &FirestoreRepository{
		client: client,
	}
}

// Create new chat in Firestore
func (r *FirestoreRepository) NewChat(ctx context.Context, chatID int64) error {
	chat := models.Chat{
		ID:            chatID,
		UserIDs:       []int64{},
		QuestionQueue: []models.QueueEntry{},
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err := r.client.Collection(chatsCollection).Doc(fmt.Sprintf("%d", chat.ID)).Set(ctx, chat)
	if err != nil {
		return fmt.Errorf("failed to save chat: %w", err)
	}
	return nil
}

// SaveChat saves or updates a chat to Firestore
func (r *FirestoreRepository) SaveChat(ctx context.Context, chat models.Chat) error {
	chat.UpdatedAt = time.Now()
	if chat.CreatedAt.IsZero() {
		chat.CreatedAt = time.Now()
	}

	_, err := r.client.Collection(chatsCollection).Doc(fmt.Sprintf("%d", chat.ID)).Set(ctx, chat)
	if err != nil {
		return fmt.Errorf("failed to save chat: %w", err)
	}
	return nil
}

// GetChat retrieves a chat by its ID
func (r *FirestoreRepository) GetChat(ctx context.Context, chatID int64) (*models.Chat, error) {
	doc, err := r.client.Collection(chatsCollection).Doc(fmt.Sprintf("%d", chatID)).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, status.Errorf(codes.NotFound, "chat not found")
		}
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}

	var chat models.Chat
	if err := doc.DataTo(&chat); err != nil {
		return nil, fmt.Errorf("failed to convert chat data: %w", err)
	}
	return &chat, nil
}

// GetAllChats retrieves all chats
func (r *FirestoreRepository) GetAllChats(ctx context.Context) ([]*models.Chat, error) {
	iter := r.client.Collection(chatsCollection).Documents(ctx)
	var chats []*models.Chat

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate chats: %w", err)
		}

		var chat models.Chat
		if err := doc.DataTo(&chat); err != nil {
			return nil, fmt.Errorf("failed to convert chat data: %w", err)
		}
		chats = append(chats, &chat)
	}

	return chats, nil
}

// GetActiveChats retrieves all active chats
func (r *FirestoreRepository) GetActiveChats(ctx context.Context) ([]*models.Chat, error) {
	iter := r.client.Collection(chatsCollection).Where("is_active", "==", true).Documents(ctx)
	var chats []*models.Chat

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate active chats: %w", err)
		}

		var chat models.Chat
		if err := doc.DataTo(&chat); err != nil {
			return nil, fmt.Errorf("failed to convert chat data: %w", err)
		}
		chats = append(chats, &chat)
	}

	return chats, nil
}

// UpdateChatUsers updates the list of users in a chat
func (r *FirestoreRepository) UpdateChatUsers(ctx context.Context, chatID int64, userIDs []int64) error {
	updates := []firestore.Update{
		{Path: "user_ids", Value: userIDs},
		{Path: "updated_at", Value: time.Now()},
	}

	_, err := r.client.Collection(chatsCollection).Doc(fmt.Sprintf("%d", chatID)).Update(ctx, updates)
	if err != nil {
		return fmt.Errorf("failed to update chat users: %w", err)
	}
	return nil
}

// AddUserToChat adds a user to the chat's user list
func (r *FirestoreRepository) AddUserToChat(ctx context.Context, chatID int64, userID int64) error {
	chat, err := r.GetChat(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %w", err)
	}

	// Check if user already exists
	for _, id := range chat.UserIDs {
		if id == userID {
			return nil // User already in chat
		}
	}

	chat.UserIDs = append(chat.UserIDs, userID)
	chat.UpdatedAt = time.Now()

	return r.SaveChat(ctx, *chat)
}

// RemoveUserFromChat removes a user from the chat's user list
func (r *FirestoreRepository) RemoveUserFromChat(ctx context.Context, chatID int64, userID int64) error {
	chat, err := r.GetChat(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %w", err)
	}

	// Remove user from user list
	var newUserIDs []int64
	for _, id := range chat.UserIDs {
		if id != userID {
			newUserIDs = append(newUserIDs, id)
		}
	}

	// Remove user from queue as well
	var newQueue []models.QueueEntry
	for _, entry := range chat.QuestionQueue {
		if entry.UserID != userID {
			newQueue = append(newQueue, entry)
		}
	}

	chat.UserIDs = newUserIDs
	chat.QuestionQueue = newQueue
	chat.UpdatedAt = time.Now()

	return r.SaveChat(ctx, *chat)
}

// UpdateQuestionQueue updates the entire question queue for a chat
func (r *FirestoreRepository) UpdateQuestionQueue(ctx context.Context, chatID int64, queue []models.QueueEntry) error {
	updates := []firestore.Update{
		{Path: "question_queue", Value: queue},
		{Path: "updated_at", Value: time.Now()},
	}

	_, err := r.client.Collection(chatsCollection).Doc(fmt.Sprintf("%d", chatID)).Update(ctx, updates)
	if err != nil {
		return fmt.Errorf("failed to update question queue: %w", err)
	}
	return nil
}

// AddToQueue adds a user to the question queue
func (r *FirestoreRepository) AddToQueue(ctx context.Context, chatID int64, userID int64) error {
	chat, err := r.GetChat(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %w", err)
	}

	// Check if user already in queue
	for _, entry := range chat.QuestionQueue {
		if entry.UserID == userID && entry.Status == models.QueueStatusWaiting {
			return nil // User already in queue
		}
	}

	// Add to end of queue
	newEntry := models.QueueEntry{
		UserID:     userID,
		Position:   len(chat.QuestionQueue),
		EnqueuedAt: time.Now(),
		Status:     models.QueueStatusWaiting,
	}

	chat.QuestionQueue = append(chat.QuestionQueue, newEntry)
	chat.UpdatedAt = time.Now()

	return r.SaveChat(ctx, *chat)
}

// RemoveFromQueue removes a user from the question queue
func (r *FirestoreRepository) RemoveFromQueue(ctx context.Context, chatID int64, userID int64) error {
	chat, err := r.GetChat(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %w", err)
	}

	var newQueue []models.QueueEntry
	for i, entry := range chat.QuestionQueue {
		if entry.UserID != userID {
			// Update positions for remaining entries
			if entry.Position > i {
				entry.Position = len(newQueue)
			}
			newQueue = append(newQueue, entry)
		}
	}

	chat.QuestionQueue = newQueue
	chat.UpdatedAt = time.Now()

	return r.SaveChat(ctx, *chat)
}

// GetNextInQueue gets the next user in the question queue
func (r *FirestoreRepository) GetNextInQueue(ctx context.Context, chatID int64) (*models.QueueEntry, error) {
	chat, err := r.GetChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}

	for _, entry := range chat.QuestionQueue {
		if entry.Status == models.QueueStatusWaiting {
			return &entry, nil
		}
	}

	return nil, status.Errorf(codes.NotFound, "no users waiting in queue")
}

// UpdateQueueEntry updates a specific queue entry
func (r *FirestoreRepository) UpdateQueueEntry(ctx context.Context, chatID int64, entry models.QueueEntry) error {
	chat, err := r.GetChat(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %w", err)
	}

	// Find and update the entry
	for i, queueEntry := range chat.QuestionQueue {
		if queueEntry.UserID == entry.UserID {
			chat.QuestionQueue[i] = entry
			chat.UpdatedAt = time.Now()
			return r.SaveChat(ctx, *chat)
		}
	}

	return status.Errorf(codes.NotFound, "queue entry not found for user")
}

// GetQueuePosition gets a user's current position in the queue
func (r *FirestoreRepository) GetQueuePosition(ctx context.Context, chatID int64, userID int64) (int, error) {
	chat, err := r.GetChat(ctx, chatID)
	if err != nil {
		return -1, fmt.Errorf("failed to get chat: %w", err)
	}

	waitingCount := 0
	for _, entry := range chat.QuestionQueue {
		if entry.Status == models.QueueStatusWaiting {
			if entry.UserID == userID {
				return waitingCount, nil
			}
			waitingCount++
		}
	}

	return -1, status.Errorf(codes.NotFound, "user not found in queue")
}

// ClearCompletedQueue removes completed entries from the queue
func (r *FirestoreRepository) ClearCompletedQueue(ctx context.Context, chatID int64) error {
	chat, err := r.GetChat(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %w", err)
	}

	var newQueue []models.QueueEntry
	position := 0
	for _, entry := range chat.QuestionQueue {
		if entry.Status == models.QueueStatusWaiting || entry.Status == models.QueueStatusAsking {
			entry.Position = position
			newQueue = append(newQueue, entry)
			position++
		}
	}

	chat.QuestionQueue = newQueue
	chat.UpdatedAt = time.Now()

	return r.SaveChat(ctx, *chat)
}

// ResetQueue clears the entire queue and rebuilds it from active users
func (r *FirestoreRepository) ResetQueue(ctx context.Context, chatID int64) error {
	chat, err := r.GetChat(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %w", err)
	}

	// Clear the queue and rebuild from active users
	var newQueue []models.QueueEntry
	for i, userID := range chat.UserIDs {
		entry := models.QueueEntry{
			UserID:     userID,
			Position:   i,
			EnqueuedAt: time.Now(),
			Status:     models.QueueStatusWaiting,
		}
		newQueue = append(newQueue, entry)
	}

	chat.QuestionQueue = newQueue
	chat.UpdatedAt = time.Now()

	return r.SaveChat(ctx, *chat)
}

// SaveChatSettings saves chat settings
func (r *FirestoreRepository) SaveChatSettings(ctx context.Context, settings models.ChatSettings) error {
	settings.UpdatedAt = time.Now()

	_, err := r.client.Collection(settingsCollection).Doc(fmt.Sprintf("%d", settings.ChatID)).Set(ctx, settings)
	if err != nil {
		return fmt.Errorf("failed to save chat settings: %w", err)
	}
	return nil
}

// GetChatSettings retrieves chat settings
func (r *FirestoreRepository) GetChatSettings(ctx context.Context, chatID int64) (*models.ChatSettings, error) {
	doc, err := r.client.Collection(settingsCollection).Doc(fmt.Sprintf("%d", chatID)).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// Return default settings if not found
			return models.DefaultChatSettings(chatID), nil
		}
		return nil, fmt.Errorf("failed to get chat settings: %w", err)
	}

	var settings models.ChatSettings
	if err := doc.DataTo(&settings); err != nil {
		return nil, fmt.Errorf("failed to convert chat settings data: %w", err)
	}
	return &settings, nil
}

// SetChatActive sets the active status of a chat
func (r *FirestoreRepository) SetChatActive(ctx context.Context, chatID int64, isActive bool) error {
	updates := []firestore.Update{
		{Path: "is_active", Value: isActive},
		{Path: "updated_at", Value: time.Now()},
	}

	_, err := r.client.Collection(chatsCollection).Doc(fmt.Sprintf("%d", chatID)).Update(ctx, updates)
	if err != nil {
		return fmt.Errorf("failed to update chat active status: %w", err)
	}
	return nil
}

// GetChatsByUser retrieves all chats that contain a specific user
func (r *FirestoreRepository) GetChatsByUser(ctx context.Context, userID int64) ([]*models.Chat, error) {
	iter := r.client.Collection(chatsCollection).Where("user_ids", "array-contains", userID).Documents(ctx)
	var chats []*models.Chat

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate chats by user: %w", err)
		}

		var chat models.Chat
		if err := doc.DataTo(&chat); err != nil {
			return nil, fmt.Errorf("failed to convert chat data: %w", err)
		}
		chats = append(chats, &chat)
	}

	return chats, nil
}
