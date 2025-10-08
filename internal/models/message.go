package models

import (
	"time"

	"github.com/go-telegram/bot/models"
)

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

func NewMessageFromTelegramUpdate(update *models.Update) *Message {
	if update.Message == nil {
		return nil
	}

	msg := update.Message
	message := &Message{
		ID:     msg.ID,
		ChatID: msg.Chat.ID,
		Date:   time.Unix(int64(msg.Date), 0),
		Text:   msg.Text,
	}

	if msg.From != nil {
		message.UserID = msg.From.ID
		message.Username = msg.From.Username
		message.FirstName = msg.From.FirstName
		message.LastName = msg.From.LastName
		message.IsBot = msg.From.IsBot
	}

	return message
}
