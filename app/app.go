package app

import (
	"log/slog"

	clients "github.com/kriku/kpukbot/internal/clients/telegram"
	repositories "github.com/kriku/kpukbot/internal/repository/messages"
	services "github.com/kriku/kpukbot/internal/services/messages"
)

type App struct {
	Logger             *slog.Logger
	MessengerClient    clients.MessengerClient
	MessagesRepository repositories.MessagesRepository
	MessagesService    *services.TelegramMessagesService
}

func NewApp(
	lo *slog.Logger,
	mc clients.MessengerClient,
	mr repositories.MessagesRepository,
	ms *services.TelegramMessagesService,
) App {
	return App{
		Logger:             lo,
		MessengerClient:    mc,
		MessagesRepository: mr,
		MessagesService:    ms,
	}
}
