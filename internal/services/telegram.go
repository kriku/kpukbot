package services

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/kriku/kpukbot/internal/config"
	"github.com/kriku/kpukbot/internal/models"
	"github.com/kriku/kpukbot/internal/repository"
)

// TelegramService handles interactions with the Telegram API
type TelegramService struct {
	Bot        *bot.Bot
	AIService  AIServiceInterface
	Repository repository.Repository
}

// NewTelegramService creates a new TelegramService
func NewTelegramService(
	c *config.Config,
	ai AIServiceInterface,
	repo repository.Repository,
) (*TelegramService, error) {
	b, err := bot.New(c.TelegramToken)
	if err != nil {
		return &TelegramService{}, err
	}

	return &TelegramService{
		Bot:        b,
		AIService:  ai,
		Repository: repo,
	}, nil
}

// HandleUpdate processes a Telegram update
func (s *TelegramService) HandleUpdate(ctx context.Context, update *tgmodels.Update) {
	if update.Message.ReplyToMessage != nil {
		log.Printf("Message is a reply to message ID: %d", update.Message.ReplyToMessage.ID)

	}

	if update.Message.From == nil {
		log.Printf("Message has no sender information")
		return
	}

	log.Printf("Handle new message: [%s] %s", update.Message.From.Username, update.Message.Text)

	user, err := s.Repository.GetUser(update.Message.From.ID)

	if err != nil {
		log.Printf("Error getting user: %s", err)
		user = &models.User{
			ID:        update.Message.From.ID,
			Username:  update.Message.From.Username,
			FirstName: update.Message.From.FirstName,
			LastName:  update.Message.From.LastName,
		}
		err = s.Repository.SaveUser(*user)
		if err != nil {
			log.Printf("Error saving user: %s", err)
		} else {
			log.Printf("User saved: %s", update.Message.From.Username)
		}
	}

	thread, err := s.Repository.GetThread(update.Message.Chat.ID)

	if err != nil {
		log.Printf("Error getting thread: %s", err)

		thread = &models.Thread{
			ID:     int64(update.Message.Chat.ID),
			ChatID: update.Message.Chat.ID,
			Messages: []models.Message{
				{
					ID:      int64(update.Message.ID),
					Sender:  *user,
					Content: update.Message.Text,
				},
			},
		}
		err = s.Repository.SaveThread(*thread)

		if err != nil {
			log.Printf("Error saving thread: %v", err)
		} else {
			log.Printf("Thread saved: %v", thread.ID)
		}
	} else {
		log.Printf("Thread already exists: %v", thread.ID)

		thread.Messages = append(thread.Messages, models.Message{
			ID:      int64(update.Message.ID),
			Sender:  *user,
			Content: update.Message.Text,
		})

		err = s.Repository.SaveThread(*thread)
		if err != nil {
			log.Printf("Error saving thread: %v", err)
		} else {
			log.Printf("Thread updated: %s", thread.ID)
		}
	}

	_, err = s.Bot.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: update.Message.Chat.ID,
		Action: tgmodels.ChatActionTyping,
	})

	if err != nil {
		log.Printf("Error sending chat typing: %s", err)
	}

	allMessages := ""
	for _, msg := range thread.Messages {
		allMessages += msg.Content + "\n"
	}

	log.Printf("All messages: %s", allMessages)

	response, err := s.AIService.GenerateContent(ctx, allMessages)
	if err != nil {
		log.Printf("Error generating content: %s", err)
		response = "Sorry, I encountered an error generating a response."
	}

	log.Printf("Generated content: %s", response)

	_, err = s.Bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      bot.EscapeMarkdown(response),
		ParseMode: tgmodels.ParseModeMarkdown,
	})

	if err != nil {
		log.Printf("Error sending message: %s", err)
	}
}
