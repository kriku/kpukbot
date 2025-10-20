package models

import "time"

type User struct {
	ID                int64              `firestore:"id"`
	FirstName         string             `firestore:"first_name"`
	LastName          string             `firestore:"last_name"`
	Username          string             `firestore:"username"`
	Bio               string             `firestore:"bio"`
	Interests         []string           `firestore:"interests"`
	Hobbies           []string           `firestore:"hobbies"`
	ChatID            int64              `firestore:"chat_id"`
	PreviousQuestions []PreviousQuestion `firestore:"previous_questions"`
	CreatedAt         time.Time          `firestore:"created_at"`
	UpdatedAt         time.Time          `firestore:"updated_at"`
}

// PreviousQuestion represents a question that was previously asked to the user
type PreviousQuestion struct {
	QuestionText string    `firestore:"question_text"`
	AskedAt      time.Time `firestore:"asked_at"`
	MessageID    int64     `firestore:"message_id"` // Reference to the original question message
}

// UserInformation represents extracted information from user introduction messages
type UserInformation struct {
	Bio       string   `json:"bio"`
	Interests []string `json:"interests"`
	Hobbies   []string `json:"hobbies"`
}

// IntroductionAnalysisResult represents the result of LLM analysis for introduction detection
type IntroductionAnalysisResult struct {
	IsIntroduction bool    `json:"is_introduction"`
	Confidence     float64 `json:"confidence"`
	Reasoning      string  `json:"reasoning"`
}
