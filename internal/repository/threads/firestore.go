package threads

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/kriku/kpukbot/internal/models"
	"google.golang.org/api/iterator"
)

type FirestoreThreadsRepository struct {
	client *firestore.Client
}

func NewFirestoreThreadsRepository(client *firestore.Client) *FirestoreThreadsRepository {
	return &FirestoreThreadsRepository{
		client: client,
	}
}

func (r *FirestoreThreadsRepository) SaveThread(ctx context.Context, thread *models.Thread) error {
	_, err := r.client.Collection("threads").Doc(thread.ID).Set(ctx, thread)
	if err != nil {
		return fmt.Errorf("failed to save thread: %w", err)
	}
	return nil
}

func (r *FirestoreThreadsRepository) GetThread(ctx context.Context, id string) (*models.Thread, error) {
	doc, err := r.client.Collection("threads").Doc(id).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread: %w", err)
	}

	var thread models.Thread
	if err := doc.DataTo(&thread); err != nil {
		return nil, fmt.Errorf("failed to unmarshal thread: %w", err)
	}

	return &thread, nil
}

func (r *FirestoreThreadsRepository) GetThreadByMessageID(ctx context.Context, messageID int) (*models.Thread, error) {
	iter := r.client.Collection("threads").
		Where("message_ids", "array-contains", messageID).
		Limit(1).
		Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to iterate threads: %w", err)
	}

	var thread models.Thread
	if err := doc.DataTo(&thread); err != nil {
		return nil, fmt.Errorf("failed to unmarshal thread: %w", err)
	}

	return &thread, nil
}

func (r *FirestoreThreadsRepository) GetThreadsByChatID(ctx context.Context, chatID int64) ([]*models.Thread, error) {
	iter := r.client.Collection("threads").
		Where("chat_id", "==", chatID).
		OrderBy("updated_at", firestore.Desc).
		Documents(ctx)

	var threads []*models.Thread
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate threads: %w", err)
		}

		var thread models.Thread
		if err := doc.DataTo(&thread); err != nil {
			return nil, fmt.Errorf("failed to unmarshal thread: %w", err)
		}
		threads = append(threads, &thread)
	}

	return threads, nil
}

func (r *FirestoreThreadsRepository) GetActiveThreadsByChatID(ctx context.Context, chatID int64) ([]*models.Thread, error) {
	iter := r.client.Collection("threads").
		Where("chat_id", "==", chatID).
		Where("is_active", "==", true).
		OrderBy("updated_at", firestore.Desc).
		Limit(10). // Limit to recent active threads
		Documents(ctx)

	var threads []*models.Thread
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate active threads: %w", err)
		}

		var thread models.Thread
		if err := doc.DataTo(&thread); err != nil {
			return nil, fmt.Errorf("failed to unmarshal thread: %w", err)
		}
		threads = append(threads, &thread)
	}

	return threads, nil
}

func (r *FirestoreThreadsRepository) UpdateThread(ctx context.Context, thread *models.Thread) error {
	_, err := r.client.Collection("threads").Doc(thread.ID).Set(ctx, thread)
	if err != nil {
		return fmt.Errorf("failed to update thread: %w", err)
	}
	return nil
}

func (r *FirestoreThreadsRepository) Close() error {
	return r.client.Close()
}
