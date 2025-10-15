package users

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/kriku/kpukbot/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUsersRepository is a mock implementation of UsersRepository for testing
type MockUsersRepository struct {
	mock.Mock
}

func (m *MockUsersRepository) SaveUser(ctx context.Context, user models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUsersRepository) GetUser(ctx context.Context, userID int64) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUsersRepository) GetUsersByChatID(ctx context.Context, chatID int64) ([]*models.User, error) {
	args := m.Called(ctx, chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUsersRepository) UpdateUserBio(ctx context.Context, userID int64, bio string) error {
	args := m.Called(ctx, userID, bio)
	return args.Error(0)
}

func (m *MockUsersRepository) UpdateUserInterests(ctx context.Context, userID int64, interests []string) error {
	args := m.Called(ctx, userID, interests)
	return args.Error(0)
}

func (m *MockUsersRepository) UpdateUserHobbies(ctx context.Context, userID int64, hobbies []string) error {
	args := m.Called(ctx, userID, hobbies)
	return args.Error(0)
}

func (m *MockUsersRepository) SearchUsersByInterest(ctx context.Context, interest string) ([]*models.User, error) {
	args := m.Called(ctx, interest)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUsersRepository) SearchUsersByHobby(ctx context.Context, hobby string) ([]*models.User, error) {
	args := m.Called(ctx, hobby)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func TestUsersService_CreateOrUpdateUser(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUsersRepository)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	service := NewUsersService(mockRepo, logger)

	userID := int64(123456)
	chatID := int64(789)
	firstName := "John"
	lastName := "Doe"
	username := "johndoe"

	// Test creating new user
	mockRepo.On("GetUser", ctx, userID).Return(nil, assert.AnError).Once()
	mockRepo.On("SaveUser", ctx, mock.MatchedBy(func(user models.User) bool {
		return user.ID == userID &&
			user.FirstName == firstName &&
			user.LastName == lastName &&
			user.Username == username &&
			user.ChatID == chatID
	})).Return(nil).Once()

	err := service.CreateOrUpdateUser(ctx, userID, chatID, firstName, lastName, username)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUsersService_AddUserInterests(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUsersRepository)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	service := NewUsersService(mockRepo, logger)

	userID := int64(123456)
	existingUser := &models.User{
		ID:        userID,
		Interests: []string{"programming", "music"},
	}
	newInterests := []string{"Gaming", "MUSIC", "  Art  "}

	mockRepo.On("GetUser", ctx, userID).Return(existingUser, nil).Once()
	mockRepo.On("UpdateUserInterests", ctx, userID, mock.MatchedBy(func(interests []string) bool {
		// Should contain deduplicated, lowercase, trimmed interests
		expectedInterests := map[string]bool{
			"programming": true,
			"music":       true,
			"gaming":      true,
			"art":         true,
		}

		if len(interests) != len(expectedInterests) {
			return false
		}

		for _, interest := range interests {
			if !expectedInterests[interest] {
				return false
			}
		}
		return true
	})).Return(nil).Once()

	err := service.AddUserInterests(ctx, userID, newInterests)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUsersService_GetUserSummary(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUsersRepository)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	service := NewUsersService(mockRepo, logger)

	userID := int64(123456)
	user := &models.User{
		ID:        userID,
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Bio:       "Software developer and musician",
		Interests: []string{"programming", "music"},
		Hobbies:   []string{"guitar", "hiking"},
	}

	mockRepo.On("GetUser", ctx, userID).Return(user, nil).Once()

	summary, err := service.GetUserSummary(ctx, userID)
	assert.NoError(t, err)
	assert.Contains(t, summary, "John Doe")
	assert.Contains(t, summary, "@johndoe")
	assert.Contains(t, summary, "Software developer and musician")
	assert.Contains(t, summary, "programming, music")
	assert.Contains(t, summary, "guitar, hiking")

	mockRepo.AssertExpectations(t)
}
