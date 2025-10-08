package models

import "time"

type Thread struct {
	ID          string    `firestore:"id"`
	ChatID      int64     `firestore:"chat_id"`
	Theme       string    `firestore:"theme"`       // Main theme/topic of the thread
	Summary     string    `firestore:"summary"`     // Brief summary of the thread
	MessageIDs  []int     `firestore:"message_ids"` // IDs of messages in this thread
	CreatedAt   time.Time `firestore:"created_at"`
	UpdatedAt   time.Time `firestore:"updated_at"`
	IsActive    bool      `firestore:"is_active"` // Whether thread is still active
	Probability float64   `firestore:"-"`         // Matching probability (not stored)
}

// ThreadMatch represents a match between a message and a thread
type ThreadMatch struct {
	Thread      *Thread
	Probability float64 // 0.0 to 1.0
	Reasoning   string  // Why this match was made
}
