package users

import (
	"context"

	"github.com/kriku/kpukbot/internal/models"
)

type UsersRepository interface {
	// SaveUser saves or updates a user in the repository
	SaveUser(ctx context.Context, user models.User) error

	// GetUser retrieves a user by their Telegram ID
	GetUser(ctx context.Context, userID int64) (*models.User, error)

	// GetUsersByChatID retrieves all users from a specific chat
	GetUsersByChatID(ctx context.Context, chatID int64) ([]*models.User, error)

	// UpdateUserBio updates the bio/introduction text for a user
	UpdateUserBio(ctx context.Context, userID int64, bio string) error

	// UpdateUserInterests updates or adds interests for a user
	UpdateUserInterests(ctx context.Context, userID int64, interests []string) error

	// UpdateUserHobbies updates or adds hobbies for a user
	UpdateUserHobbies(ctx context.Context, userID int64, hobbies []string) error

	// SearchUsersByInterest finds users who have specific interests
	SearchUsersByInterest(ctx context.Context, interest string) ([]*models.User, error)

	// SearchUsersByHobby finds users who have specific hobbies
	SearchUsersByHobby(ctx context.Context, hobby string) ([]*models.User, error)
}
