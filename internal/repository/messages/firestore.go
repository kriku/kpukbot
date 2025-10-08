package messages

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/kriku/kpukbot/internal/models"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	messagesCollection = "messages"
)

// FirestoreRepository implements Repository interface using Firestore
type FirestoreRepository struct {
	client *firestore.Client
}

// NewFirestoreMessagesRepository creates a new FirestoreRepository with existing client
func NewFirestoreMessagesRepository(client *firestore.Client) MessagesRepository {
	return &FirestoreRepository{
		client: client,
	}
}

// SaveMessage saves a message to Firestore
func (r *FirestoreRepository) SaveMessage(ctx context.Context, m models.Message) error {
	_, err := r.client.Collection(messagesCollection).Doc(fmt.Sprintf("%d", m.ID)).Set(ctx, m)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}
	return nil
}

func (r *FirestoreRepository) GetMessages(ctx context.Context, chatID int64) ([]*models.Message, error) {
	iter := r.client.Collection(messagesCollection).
		Where("chat_id", "==", chatID).
		Documents(ctx)
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
		return nil, status.Errorf(codes.NotFound, "no messages found")
	}

	return messages, nil
}

func (r *FirestoreRepository) GetMessage(ctx context.Context, id int64) ([]*models.Message, error) {
	iter := r.client.Collection(messagesCollection).Where("id", "==", id).Documents(ctx)
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
