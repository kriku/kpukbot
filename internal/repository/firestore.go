package repository

import (
	"context"
	"strconv"

	"cloud.google.com/go/firestore"
	"github.com/kriku/kpukbot/internal/config"
	"github.com/kriku/kpukbot/internal/models"
	"google.golang.org/api/iterator"
)

const (
	// CollectionUsers is the Firestore collection for users
	CollectionUsers = "users"
	// CollectionThreads is the Firestore collection for threads
	CollectionThreads = "threads"
)

// FirestoreRepository implements the Repository interface using Google Firestore
type FirestoreRepository struct {
	client *firestore.Client
	ctx    context.Context
}

// NewFirestoreRepository creates a new Firestore repository
func NewFirestoreRepository(c *config.Config) (Repository, error) {
	ctx := context.Background()

	client, err := firestore.NewClient(ctx, c.FilestoreConfig.ProjectID)
	if err != nil {
		return nil, err
	}

	return &FirestoreRepository{
		client: client,
		ctx:    ctx,
	}, nil
}

// SaveUser stores user data in Firestore
func (r *FirestoreRepository) SaveUser(u models.User) error {
	_, err := r.client.Collection("users").Doc(strconv.FormatInt(u.ID, 10)).Set(r.ctx, u)
	return err
}

// GetUser retrieves user data from Firestore
func (r *FirestoreRepository) GetUser(userID int64) (*models.User, error) {
	doc, err := r.client.Collection("users").Doc(strconv.FormatInt(userID, 10)).Get(r.ctx)

	if err != nil {
		return nil, err
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// SaveThread stores chat data in Firestore
func (r *FirestoreRepository) SaveThread(t models.Thread) error {
	_, err := r.client.Collection(CollectionThreads).Doc(strconv.FormatInt(t.ID, 10)).Set(r.ctx, t)
	return err
}

// GetThread retrieves chat data from Firestore
func (r *FirestoreRepository) GetThread(threadID int64) (*models.Thread, error) {
	doc, err := r.client.Collection(CollectionThreads).Doc(strconv.FormatInt(threadID, 10)).Get(r.ctx)
	if err != nil {
		return nil, err
	}

	var thread models.Thread
	if err := doc.DataTo(&thread); err != nil {
		return nil, err
	}

	return &thread, nil
}

// SaveMessage stores message data in Firestore
func (r *FirestoreRepository) SaveMessage(threadId int64, m models.Message) error {
	_, err := r.client.Collection(CollectionThreads).Doc(strconv.FormatInt(threadId, 10)).Collection("messages").Doc(string(m.ID)).Set(r.ctx, m)
	return err
}

// GetMessages retrieves message history from Firestore
func (r *FirestoreRepository) GetMessages(threadId int64) ([]*models.Message, error) {
	query := r.client.Collection(CollectionThreads).Doc(strconv.FormatInt(threadId, 10)).Collection("messages")

	iter := query.Documents(r.ctx)
	defer iter.Stop()

	var messages []*models.Message
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var message models.Message
		if err := doc.DataTo(&message); err != nil {
			return nil, err
		}
		messages = append(messages, &message)
	}

	return messages, nil
}

// Close terminates the connection to Firestore
func (r *FirestoreRepository) Close() error {
	return r.client.Close()
}
