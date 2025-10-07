package messages

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/kriku/kpukbot/internal/config"
	"github.com/kriku/kpukbot/internal/models"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	messagesCollection = "messages"
	defaultTimeout     = 30 * time.Second
)

// FirestoreRepository implements Repository interface using Firestore
type FirestoreRepository struct {
	client *firestore.Client
	ctx    context.Context
}

// NewFirestoreRepository creates a new FirestoreRepository
func NewFirestoreRepository(c *config.Config) (MessagesRepository, error) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, c.FilestoreConfig.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client: %w", err)
	}

	return &FirestoreRepository{
		client: client,
		ctx:    ctx,
	}, nil
}

// SaveMessage saves a message to Firestore
func (r *FirestoreRepository) SaveMessage(ctx context.Context, m models.Message) error {
	// ctx, cancel := context.WithTimeout(r.ctx, defaultTimeout)
	// defer cancel()

	_, err := r.client.Collection(messagesCollection).Doc(fmt.Sprintf("%d", m.ID)).Set(ctx, m)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	return nil
}

// GetMessage retrieves messages from Firestore based on ID
func (r *FirestoreRepository) GetMessage(ctx context.Context, id int64) ([]*models.Message, error) {
	// ctx, cancel := context.WithTimeout(r.ctx, defaultTimeout)
	// defer cancel()

	iter := r.client.Collection(messagesCollection).Where("ID", "==", id).Documents(ctx)
	defer iter.Stop()

	var messages []*models.Message
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate messages: %w", err)
		}

		var m models.Message
		if err := doc.DataTo(&m); err != nil {
			return nil, fmt.Errorf("failed to unmarshal message: %w", err)
		}

		messages = append(messages, &m)
	}

	if len(messages) == 0 {
		return nil, status.Errorf(codes.NotFound, "message with id %d not found", id)
	}

	return messages, nil
}

// Close closes the Firestore client connection
func (r *FirestoreRepository) Close() error {
	return r.client.Close()
}
