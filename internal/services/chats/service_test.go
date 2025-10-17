package chats

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/kriku/kpukbot/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockChatsRepository is a mock implementation of ChatsRepository for testing
type MockChatsRepository struct {
	mock.Mock
}

func (m *MockChatsRepository) NewChat(ctx context.Context, chatID int64) error {
	args := m.Called(ctx, chatID)
	return args.Error(0)
}

func (m *MockChatsRepository) SaveChat(ctx context.Context, chat models.Chat) error {
	args := m.Called(ctx, chat)
	return args.Error(0)
}

func (m *MockChatsRepository) GetChat(ctx context.Context, chatID int64) (*models.Chat, error) {
	args := m.Called(ctx, chatID)
	return args.Get(0).(*models.Chat), args.Error(1)
}

func (m *MockChatsRepository) GetAllChats(ctx context.Context) ([]*models.Chat, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Chat), args.Error(1)
}

func (m *MockChatsRepository) GetActiveChats(ctx context.Context) ([]*models.Chat, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Chat), args.Error(1)
}

func (m *MockChatsRepository) UpdateChatUsers(ctx context.Context, chatID int64, userIDs []int64) error {
	args := m.Called(ctx, chatID, userIDs)
	return args.Error(0)
}

func (m *MockChatsRepository) AddUserToChat(ctx context.Context, chatID int64, userID int64) error {
	args := m.Called(ctx, chatID, userID)
	return args.Error(0)
}

func (m *MockChatsRepository) RemoveUserFromChat(ctx context.Context, chatID int64, userID int64) error {
	args := m.Called(ctx, chatID, userID)
	return args.Error(0)
}

func (m *MockChatsRepository) UpdateQuestionQueue(ctx context.Context, chatID int64, queue []models.QueueEntry) error {
	args := m.Called(ctx, chatID, queue)
	return args.Error(0)
}

func (m *MockChatsRepository) AddToQueue(ctx context.Context, chatID int64, userID int64) error {
	args := m.Called(ctx, chatID, userID)
	return args.Error(0)
}

func (m *MockChatsRepository) RemoveFromQueue(ctx context.Context, chatID int64, userID int64) error {
	args := m.Called(ctx, chatID, userID)
	return args.Error(0)
}

func (m *MockChatsRepository) GetNextInQueue(ctx context.Context, chatID int64) (*models.QueueEntry, error) {
	args := m.Called(ctx, chatID)
	return args.Get(0).(*models.QueueEntry), args.Error(1)
}

func (m *MockChatsRepository) UpdateQueueEntry(ctx context.Context, chatID int64, entry models.QueueEntry) error {
	args := m.Called(ctx, chatID, entry)
	return args.Error(0)
}

func (m *MockChatsRepository) GetQueuePosition(ctx context.Context, chatID int64, userID int64) (int, error) {
	args := m.Called(ctx, chatID, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockChatsRepository) ClearCompletedQueue(ctx context.Context, chatID int64) error {
	args := m.Called(ctx, chatID)
	return args.Error(0)
}

func (m *MockChatsRepository) ResetQueue(ctx context.Context, chatID int64) error {
	args := m.Called(ctx, chatID)
	return args.Error(0)
}

func (m *MockChatsRepository) SaveChatSettings(ctx context.Context, settings models.ChatSettings) error {
	args := m.Called(ctx, settings)
	return args.Error(0)
}

func (m *MockChatsRepository) GetChatSettings(ctx context.Context, chatID int64) (*models.ChatSettings, error) {
	args := m.Called(ctx, chatID)
	return args.Get(0).(*models.ChatSettings), args.Error(1)
}

func (m *MockChatsRepository) SetChatActive(ctx context.Context, chatID int64, isActive bool) error {
	args := m.Called(ctx, chatID, isActive)
	return args.Error(0)
}

func (m *MockChatsRepository) GetChatsByUser(ctx context.Context, userID int64) ([]*models.Chat, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.Chat), args.Error(1)
}

func TestChatsService_EnqueueUser(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockChatsRepository)
	logger := slog.Default()
	service := NewChatsService(mockRepo, logger)

	chatID := int64(123)
	userID := int64(456)

	// Mock chat settings
	settings := &models.ChatSettings{
		ChatID:               chatID,
		EnableQuestionRounds: true,
	}

	mockRepo.On("GetChatSettings", ctx, chatID).Return(settings, nil)
	mockRepo.On("AddToQueue", ctx, chatID, userID).Return(nil)

	err := service.EnqueueUser(ctx, chatID, userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestChatsService_GetNextUserInQueue(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockChatsRepository)
	logger := slog.Default()
	service := NewChatsService(mockRepo, logger)

	chatID := int64(123)
	userID := int64(456)

	expectedEntry := &models.QueueEntry{
		UserID:     userID,
		Position:   0,
		EnqueuedAt: time.Now(),
		Status:     models.QueueStatusWaiting,
	}

	mockRepo.On("GetNextInQueue", ctx, chatID).Return(expectedEntry, nil)

	entry, err := service.GetNextUserInQueue(ctx, chatID)

	assert.NoError(t, err)
	assert.Equal(t, expectedEntry, entry)
	assert.Equal(t, userID, entry.UserID)
	assert.Equal(t, models.QueueStatusWaiting, entry.Status)
	mockRepo.AssertExpectations(t)
}

func TestChatsService_MarkQuestionAsked(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockChatsRepository)
	logger := slog.Default()
	service := NewChatsService(mockRepo, logger)

	chatID := int64(123)
	userID := int64(456)
	questionID := "question_123"

	mockRepo.On("UpdateQueueEntry", ctx, chatID, mock.MatchedBy(func(entry models.QueueEntry) bool {
		return entry.UserID == userID &&
			entry.Status == models.QueueStatusAsking &&
			entry.QuestionID == questionID &&
			entry.AskedAt != nil
	})).Return(nil)

	err := service.MarkQuestionAsked(ctx, chatID, userID, questionID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
