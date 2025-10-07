package models

import "time"

type Message struct {
	ID        int       `firestore:"id"`
	ChatID    int64     `firestore:"chat_id"`
	UserID    int64     `firestore:"user_id"`
	Text      string    `firestore:"text"`
	Username  string    `firestore:"username"`
	FirstName string    `firestore:"first_name"`
	LastName  string    `firestore:"last_name"`
	Date      time.Time `firestore:"date"`
	IsBot     bool      `firestore:"is_bot"`
	// ChatType  string    `firestore:"chat_type"`
}
