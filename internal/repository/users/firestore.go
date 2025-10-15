package users

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
	usersCollection = "users"
)

// FirestoreRepository implements UsersRepository interface using Firestore
type FirestoreRepository struct {
	client *firestore.Client
}

// NewFirestoreUsersRepository creates a new FirestoreRepository with existing client
func NewFirestoreUsersRepository(client *firestore.Client) UsersRepository {
	return &FirestoreRepository{
		client: client,
	}
}

// SaveUser saves or updates a user to Firestore
func (r *FirestoreRepository) SaveUser(ctx context.Context, user models.User) error {
	user.UpdatedAt = time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	_, err := r.client.Collection(usersCollection).Doc(fmt.Sprintf("%d", user.ID)).Set(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}
	return nil
}

// GetUser retrieves a user by their Telegram ID
func (r *FirestoreRepository) GetUser(ctx context.Context, userID int64) (*models.User, error) {
	doc, err := r.client.Collection(usersCollection).Doc(fmt.Sprintf("%d", userID)).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

// GetUsersByChatID retrieves all users from a specific chat
func (r *FirestoreRepository) GetUsersByChatID(ctx context.Context, chatID int64) ([]*models.User, error) {
	iter := r.client.Collection(usersCollection).
		Where("chat_id", "==", chatID).
		Documents(ctx)
	defer iter.Stop()

	var users []*models.User
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate users: %w", err)
		}

		var user models.User
		if err := doc.DataTo(&user); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// UpdateUserBio updates the bio/introduction text for a user
func (r *FirestoreRepository) UpdateUserBio(ctx context.Context, userID int64, bio string) error {
	_, err := r.client.Collection(usersCollection).Doc(fmt.Sprintf("%d", userID)).Update(ctx, []firestore.Update{
		{
			Path:  "bio",
			Value: bio,
		},
		{
			Path:  "updated_at",
			Value: time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update user bio: %w", err)
	}
	return nil
}

// UpdateUserInterests updates or adds interests for a user
func (r *FirestoreRepository) UpdateUserInterests(ctx context.Context, userID int64, interests []string) error {
	_, err := r.client.Collection(usersCollection).Doc(fmt.Sprintf("%d", userID)).Update(ctx, []firestore.Update{
		{
			Path:  "interests",
			Value: interests,
		},
		{
			Path:  "updated_at",
			Value: time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update user interests: %w", err)
	}
	return nil
}

// UpdateUserHobbies updates or adds hobbies for a user
func (r *FirestoreRepository) UpdateUserHobbies(ctx context.Context, userID int64, hobbies []string) error {
	_, err := r.client.Collection(usersCollection).Doc(fmt.Sprintf("%d", userID)).Update(ctx, []firestore.Update{
		{
			Path:  "hobbies",
			Value: hobbies,
		},
		{
			Path:  "updated_at",
			Value: time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update user hobbies: %w", err)
	}
	return nil
}

// SearchUsersByInterest finds users who have specific interests
func (r *FirestoreRepository) SearchUsersByInterest(ctx context.Context, interest string) ([]*models.User, error) {
	iter := r.client.Collection(usersCollection).
		Where("interests", "array-contains", interest).
		Documents(ctx)
	defer iter.Stop()

	var users []*models.User
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate users: %w", err)
		}

		var user models.User
		if err := doc.DataTo(&user); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// SearchUsersByHobby finds users who have specific hobbies
func (r *FirestoreRepository) SearchUsersByHobby(ctx context.Context, hobby string) ([]*models.User, error) {
	iter := r.client.Collection(usersCollection).
		Where("hobbies", "array-contains", hobby).
		Documents(ctx)
	defer iter.Stop()

	var users []*models.User
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate users: %w", err)
		}

		var user models.User
		if err := doc.DataTo(&user); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}
