package models

import "time"

// Chat represents a chat group with its users and question queue
type Chat struct {
	ID            int64        `firestore:"id"`
	Title         string       `firestore:"title"`
	Type          string       `firestore:"type"` // group, supergroup, private
	Description   string       `firestore:"description"`
	UserIDs       []int64      `firestore:"user_ids"`       // List of user IDs in the chat
	QuestionQueue []QueueEntry `firestore:"question_queue"` // Queue of users waiting for questions
	IsActive      bool         `firestore:"is_active"`      // Whether chat is active
	CreatedAt     time.Time    `firestore:"created_at"`
	UpdatedAt     time.Time    `firestore:"updated_at"`
}

// QueueEntry represents a user's position and state in the question queue
type QueueEntry struct {
	UserID     int64      `firestore:"user_id"`
	Position   int        `firestore:"position"`    // Position in queue (0-based)
	EnqueuedAt time.Time  `firestore:"enqueued_at"` // When user was added to queue
	Status     string     `firestore:"status"`      // waiting, asking, completed, skipped
	QuestionID string     `firestore:"question_id"` // ID of the current/last question asked
	AskedAt    *time.Time `firestore:"asked_at"`    // When question was asked (nil if not asked yet)
	AnsweredAt *time.Time `firestore:"answered_at"` // When user answered (nil if not answered yet)
}

// QueueStatus represents possible queue entry statuses
const (
	QueueStatusWaiting   = "waiting"   // User is waiting in queue
	QueueStatusAsking    = "asking"    // User is currently being asked a question
	QueueStatusCompleted = "completed" // User has answered the question
	QueueStatusSkipped   = "skipped"   // User was skipped for this round
)

// ChatSettings represents configurable settings for a chat
type ChatSettings struct {
	ChatID               int64         `firestore:"chat_id"`
	QuestionInterval     time.Duration `firestore:"question_interval"`      // How often to ask questions
	MaxQueueSize         int           `firestore:"max_queue_size"`         // Maximum queue size
	AutoEnqueueNewUsers  bool          `firestore:"auto_enqueue_new_users"` // Auto-add new users to queue
	SkipInactiveUsers    bool          `firestore:"skip_inactive_users"`    // Skip users who don't respond
	InactivityTimeout    time.Duration `firestore:"inactivity_timeout"`     // How long to wait before skipping
	EnableQuestionRounds bool          `firestore:"enable_question_rounds"` // Enable question rounds feature
	UpdatedAt            time.Time     `firestore:"updated_at"`
}

// DefaultChatSettings returns default settings for a new chat
func DefaultChatSettings(chatID int64) *ChatSettings {
	return &ChatSettings{
		ChatID:               chatID,
		QuestionInterval:     24 * time.Hour, // Ask questions once per day
		MaxQueueSize:         50,
		AutoEnqueueNewUsers:  true,
		SkipInactiveUsers:    true,
		InactivityTimeout:    2 * time.Hour,
		EnableQuestionRounds: true,
		UpdatedAt:            time.Now(),
	}
}
