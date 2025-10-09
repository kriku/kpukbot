package app

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/kriku/kpukbot/internal/config"
)

func NewFirestoreClient(ctx context.Context, c *config.Config) (*firestore.Client, error) {
	client, err := firestore.NewClient(ctx, c.FilestoreConfig.ProjectID)

	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client: %w", err)
	}

	return client, nil
}
