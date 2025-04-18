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
	if update.Message == nil {
		log.Printf("Update contains no message")
		return
	}

	if update.Message.From == nil {
		log.Printf("Message has no sender information")
		return
	}

	if update.Message.ReplyToMessage != nil {
		log.Printf("Message is a reply to message ID: %d", update.Message.ReplyToMessage.ID)
	}

	log.Printf("Handle new message: [%s] %s", update.Message.From.Username, update.Message.Text)

	user := s.processUser(update.Message.From)
	thread := s.processThread(update.Message, user)
	s.sendResponse(ctx, update.Message.Chat.ID, thread)
}

// processUser gets or creates a user based on the message sender
func (s *TelegramService) processUser(from *tgmodels.User) *models.User {
	user, err := s.Repository.GetUser(from.ID)
	if err != nil {
		log.Printf("Error getting user: %s", err)
		user = &models.User{
			ID:        from.ID,
			Username:  from.Username,
			FirstName: from.FirstName,
			LastName:  from.LastName,
		}
		err = s.Repository.SaveUser(*user)
		if err != nil {
			log.Printf("Error saving user: %s", err)
		} else {
			log.Printf("User saved: %s", from.Username)
		}
	}
	return user
}

// processThread gets or creates a thread and adds the current message to it
func (s *TelegramService) processThread(message *tgmodels.Message, user *models.User) *models.Thread {
	thread, err := s.Repository.GetThread(message.Chat.ID)

	if err != nil {
		log.Printf("Error getting thread: %s", err)

		thread = &models.Thread{
			ID:     int64(message.Chat.ID),
			ChatID: message.Chat.ID,
			Messages: []models.Message{
				{
					ID:      int64(message.ID),
					Sender:  *user,
					Content: message.Text,
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
			ID:      int64(message.ID),
			Sender:  *user,
			Content: message.Text,
		})

		err = s.Repository.SaveThread(*thread)
		if err != nil {
			log.Printf("Error saving thread: %v", err)
		} else {
			log.Printf("Thread updated: %s", thread.ID)
		}
	}

	return thread
}

// sendResponse generates AI content and sends it as a response
func (s *TelegramService) sendResponse(ctx context.Context, chatID int64, thread *models.Thread) {
	_, err := s.Bot.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: chatID,
		Action: tgmodels.ChatActionTyping,
	})

	if err != nil {
		log.Printf("Error sending chat typing: %s", err)
	}

	allMessages := s.formatMessagesForAI(thread)
	log.Printf("All messages: %s", allMessages)

	response, err := s.AIService.GenerateContent(ctx, allMessages)
	if err != nil {
		log.Printf("Error generating content: %s", err)
		response = "Sorry, I encountered an error generating a response."
	}

	log.Printf("Generated content: %s", response)

	_, err = s.Bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      bot.EscapeMarkdown(response),
		ParseMode: tgmodels.ParseModeMarkdown,
	})

	if err != nil {
		log.Printf("Error sending message: %s", err)
	}
}

// formatMessagesForAI concatenates all messages in a thread for AI processing
func (s *TelegramService) formatMessagesForAI(thread *models.Thread) string {
	allMessages := ""
	for _, msg := range thread.Messages {
		allMessages += msg.Content + "\n"
	}
	return allMessages
}
