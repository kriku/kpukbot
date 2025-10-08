package app

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/kriku/kpukbot/internal/config"
)

// NewFirestoreClient creates a new Firestore client
func NewFirestoreClient(c *config.Config) (*firestore.Client, error) {
	ctx := context.Background()

	client, err := firestore.NewClient(ctx, c.FilestoreConfig.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client: %w", err)
	}

	return client, nil
}
