package models

type User struct {
	ID        int64  `firestore:"id"`
	FirstName string `firestore:"first_name"`
	LastName  string `firestore:"last_name"`
	Username  string `firestore:"username"`
}
