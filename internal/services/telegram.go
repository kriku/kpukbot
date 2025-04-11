package services

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/kriku/kpukbot/internal/config"
)

// TelegramService handles interactions with the Telegram API
type TelegramService struct {
	Bot       *bot.Bot
	AIService AIServiceInterface
}

// NewTelegramService creates a new TelegramService
func NewTelegramService(c *config.Config, aiService AIServiceInterface) (*TelegramService, error) {
	b, err := bot.New(c.TelegramToken)
	if err != nil {
		return &TelegramService{}, err
	}

	return &TelegramService{
		Bot:       b,
		AIService: aiService,
	}, nil
}

// HandleUpdate processes a Telegram update
func (s *TelegramService) HandleUpdate(ctx context.Context, update *models.Update) {
	if update.Message.From != nil {
		log.Printf("Handle new message: [%s] %s", update.Message.From.Username, update.Message.Text)
	} else {
		log.Printf("Handle new message in chat: [%s] %s", update.Message.Chat.Title, update.Message.Text)
	}

	_, err := s.Bot.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: update.Message.Chat.ID,
		Action: models.ChatActionTyping,
	})

	if err != nil {
		log.Printf("Error sending chat typing: %s", err)
	}

	response, err := s.AIService.GenerateContent(ctx, update.Message.Text)
	if err != nil {
		log.Printf("Error generating content: %s", err)
		response = "Sorry, I encountered an error generating a response."
	}

	log.Printf("Generated content: %s", response)

	_, err = s.Bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      bot.EscapeMarkdown(response),
		ParseMode: models.ParseModeMarkdown,
	})

	if err != nil {
		log.Printf("Error sending message: %s", err)
	}
}
