package telegram

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/kriku/kpukbot/internal/config"
)

type MessengerClient interface {
	Start(ctx context.Context) error
	SendMessage(ctx context.Context, chatID int64, text string) (*models.Message, error)
	HandleWebhook(ctx context.Context, res http.ResponseWriter, req *http.Request)

	Close() error
}

type TelegramClient struct {
	bot *bot.Bot
}

func NewTelegramClient(ctx context.Context, c *config.Config, handler bot.HandlerFunc) (MessengerClient, error) {
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

func (t *TelegramClient) Start(ctx context.Context) error {
	t.bot.Start(ctx)

	return nil
}

func (t *TelegramClient) SendMessage(ctx context.Context, chatID int64, text string) (*models.Message, error) {
	msg := bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	}

	message, err := t.bot.SendMessage(ctx, &msg)

	if err != nil {
		return nil, err
	}

	return message, nil
}

func (t *TelegramClient) HandleWebhook(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	update := models.Update{}
	json.NewDecoder(req.Body).Decode(&update)
	t.bot.ProcessUpdate(ctx, &update)
}

func (t *TelegramClient) Close() error {
	return nil
}
