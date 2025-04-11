package models

type Thread struct {
	ID       int64     `firestore:"id"`
	ChatID   int64     `firestore:"chat_id"`
	Messages []Message `firestore:"messages"`
}
