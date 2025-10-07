package telegram

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/kriku/kpukbot/internal/config"
)

type MessengerClient interface {
	Start() error
	SendMessage(chatID int64, text string) (*models.Message, error)
	HandleWebhook(res http.ResponseWriter, req *http.Request)

	Close() error
}

type TelegramClient struct {
	bot *bot.Bot
}

func NewTelegramClient(c *config.Config, handler bot.HandlerFunc) (MessengerClient, error) {
	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(c.TelegramToken, opts...)

	if err != nil {
		return nil, err
	}

	return &TelegramClient{
		bot: b,
	}, nil
}

func (t *TelegramClient) Start() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	t.bot.Start(ctx)

	return nil
}

func (t *TelegramClient) SendMessage(chatID int64, text string) (*models.Message, error) {
	// Implement SendMessage method
	msg := bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	}
	ctx := context.Background()
	message, err := t.bot.SendMessage(ctx, &msg)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (t *TelegramClient) HandleWebhook(res http.ResponseWriter, req *http.Request) {
	update := models.Update{}

	json.NewDecoder(req.Body).Decode(&update)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	t.bot.ProcessUpdate(ctx, &update)
}

func (t *TelegramClient) Close() error {
	// Implement Close method
	return nil
}
