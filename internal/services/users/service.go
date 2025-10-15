package users

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/kriku/kpukbot/internal/models"
	"github.com/kriku/kpukbot/internal/repository/users"
)

// UsersService provides business logic for user management
type UsersService struct {
	repository users.UsersRepository
	logger     *slog.Logger
}

// NewUsersService creates a new users service
func NewUsersService(repository users.UsersRepository, logger *slog.Logger) *UsersService {
	return &UsersService{
		repository: repository,
		logger:     logger,
	}
}

// CreateOrUpdateUser creates a new user or updates an existing one
func (s *UsersService) CreateOrUpdateUser(ctx context.Context, userID int64, chatID int64, firstName, lastName, username string) error {
	user := models.User{
		ID:        userID,
		FirstName: firstName,
		LastName:  lastName,
		Username:  username,
		ChatID:    chatID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Check if user already exists
	existingUser, err := s.repository.GetUser(ctx, userID)
	if err == nil && existingUser != nil {
		// User exists, preserve existing data and update basic info
		user.Bio = existingUser.Bio
		user.Interests = existingUser.Interests
		user.Hobbies = existingUser.Hobbies
		user.CreatedAt = existingUser.CreatedAt
	}

	err = s.repository.SaveUser(ctx, user)
	if err != nil {
		s.logger.Error("Failed to save user", "user_id", userID, "error", err)
		return fmt.Errorf("failed to save user: %w", err)
	}

	s.logger.Info("User saved successfully", "user_id", userID, "username", username)
	return nil
}

// UpdateUserBio updates the user's bio/introduction
func (s *UsersService) UpdateUserBio(ctx context.Context, userID int64, bio string) error {
	err := s.repository.UpdateUserBio(ctx, userID, bio)
	if err != nil {
		s.logger.Error("Failed to update user bio", "user_id", userID, "error", err)
		return fmt.Errorf("failed to update user bio: %w", err)
	}

	s.logger.Info("User bio updated", "user_id", userID)
	return nil
}

// AddUserInterests adds interests to a user (merges with existing)
func (s *UsersService) AddUserInterests(ctx context.Context, userID int64, newInterests []string) error {
	// Get current user to merge interests
	user, err := s.repository.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Merge and deduplicate interests
	interestsMap := make(map[string]bool)
	for _, interest := range user.Interests {
		interestsMap[strings.ToLower(strings.TrimSpace(interest))] = true
	}
	for _, interest := range newInterests {
		cleaned := strings.ToLower(strings.TrimSpace(interest))
		if cleaned != "" {
			interestsMap[cleaned] = true
		}
	}

	// Convert back to slice
	var allInterests []string
	for interest := range interestsMap {
		allInterests = append(allInterests, interest)
	}

	err = s.repository.UpdateUserInterests(ctx, userID, allInterests)
	if err != nil {
		s.logger.Error("Failed to update user interests", "user_id", userID, "error", err)
		return fmt.Errorf("failed to update user interests: %w", err)
	}

	s.logger.Info("User interests updated", "user_id", userID, "interests_count", len(allInterests))
	return nil
}

// AddUserHobbies adds hobbies to a user (merges with existing)
func (s *UsersService) AddUserHobbies(ctx context.Context, userID int64, newHobbies []string) error {
	// Get current user to merge hobbies
	user, err := s.repository.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Merge and deduplicate hobbies
	hobbiesMap := make(map[string]bool)
	for _, hobby := range user.Hobbies {
		hobbiesMap[strings.ToLower(strings.TrimSpace(hobby))] = true
	}
	for _, hobby := range newHobbies {
		cleaned := strings.ToLower(strings.TrimSpace(hobby))
		if cleaned != "" {
			hobbiesMap[cleaned] = true
		}
	}

	// Convert back to slice
	var allHobbies []string
	for hobby := range hobbiesMap {
		allHobbies = append(allHobbies, hobby)
	}

	err = s.repository.UpdateUserHobbies(ctx, userID, allHobbies)
	if err != nil {
		s.logger.Error("Failed to update user hobbies", "user_id", userID, "error", err)
		return fmt.Errorf("failed to update user hobbies: %w", err)
	}

	s.logger.Info("User hobbies updated", "user_id", userID, "hobbies_count", len(allHobbies))
	return nil
}

// GetUser retrieves a user by ID
func (s *UsersService) GetUser(ctx context.Context, userID int64) (*models.User, error) {
	return s.repository.GetUser(ctx, userID)
}

// GetChatUsers retrieves all users from a specific chat
func (s *UsersService) GetChatUsers(ctx context.Context, chatID int64) ([]*models.User, error) {
	return s.repository.GetUsersByChatID(ctx, chatID)
}

// FindUsersByInterest finds users with similar interests
func (s *UsersService) FindUsersByInterest(ctx context.Context, interest string) ([]*models.User, error) {
	return s.repository.SearchUsersByInterest(ctx, strings.ToLower(strings.TrimSpace(interest)))
}

// FindUsersByHobby finds users with similar hobbies
func (s *UsersService) FindUsersByHobby(ctx context.Context, hobby string) ([]*models.User, error) {
	return s.repository.SearchUsersByHobby(ctx, strings.ToLower(strings.TrimSpace(hobby)))
}

// GetUserSummary returns a formatted summary of a user's information
func (s *UsersService) GetUserSummary(ctx context.Context, userID int64) (string, error) {
	user, err := s.repository.GetUser(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	var summary strings.Builder

	// Basic info
	if user.FirstName != "" {
		summary.WriteString(fmt.Sprintf("Name: %s", user.FirstName))
		if user.LastName != "" {
			summary.WriteString(fmt.Sprintf(" %s", user.LastName))
		}
		summary.WriteString("\n")
	}

	if user.Username != "" {
		summary.WriteString(fmt.Sprintf("Username: @%s\n", user.Username))
	}

	// Bio
	if user.Bio != "" {
		summary.WriteString(fmt.Sprintf("Bio: %s\n", user.Bio))
	}

	// Interests
	if len(user.Interests) > 0 {
		summary.WriteString(fmt.Sprintf("Interests: %s\n", strings.Join(user.Interests, ", ")))
	}

	// Hobbies
	if len(user.Hobbies) > 0 {
		summary.WriteString(fmt.Sprintf("Hobbies: %s\n", strings.Join(user.Hobbies, ", ")))
	}

	return summary.String(), nil
}
