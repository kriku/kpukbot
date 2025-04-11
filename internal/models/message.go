package models

type Message struct {
	ID      int64  `firestore:"id"`
	Content string `firestore:"content"`
	Sender  User   `firestore:"sender"`
}
