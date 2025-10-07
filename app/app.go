package app

import (
	"log/slog"

	clients "github.com/kriku/kpukbot/internal/clients/telegram"
	repositories "github.com/kriku/kpukbot/internal/repository/messages"
)

type App struct {
	Logger             *slog.Logger
	MessengerClient    clients.MessengerClient
	MessagesRepository repositories.MessagesRepository
}

func NewApp(
	lo *slog.Logger,
	mc clients.MessengerClient,
	mr repositories.MessagesRepository,
) App {
	return App{
		Logger:             lo,
		MessengerClient:    mc,
		MessagesRepository: mr,
	}
}
