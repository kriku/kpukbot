package models

import "time"

type User struct {
	ID        int64     `firestore:"id"`
	FirstName string    `firestore:"first_name"`
	LastName  string    `firestore:"last_name"`
	Username  string    `firestore:"username"`
	Bio       string    `firestore:"bio"`
	Interests []string  `firestore:"interests"`
	Hobbies   []string  `firestore:"hobbies"`
	ChatID    int64     `firestore:"chat_id"`
	CreatedAt time.Time `firestore:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at"`
}
