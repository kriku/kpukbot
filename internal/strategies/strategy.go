package strategies

import (
	"context"

	"github.com/kriku/kpukbot/internal/models"
)

// ResponseStrategy defines the interface for different response strategies
type ResponseStrategy interface {
	// ShouldRespond determines if this strategy should respond to the message
	// Returns: shouldRespond, confidence (0.0-1.0), error
	ShouldRespond(ctx context.Context, thread *models.Thread, messages []*models.Message, newMessage *models.Message) (bool, float64, error)

	// GenerateResponse generates the actual response
	GenerateResponse(ctx context.Context, thread *models.Thread, messages []*models.Message, newMessage *models.Message) (string, error)

	// Name returns the strategy name
	Name() string

	// Priority returns the priority of this strategy (higher = more important)
	Priority() int
}

// StrategyResult holds the result of a strategy evaluation
type StrategyResult struct {
	Strategy      ResponseStrategy
	ShouldRespond bool
	Confidence    float64
	Response      string
	Error         error
}
