//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"

	clients "github.com/kriku/kpukbot/internal/clients/telegram"
	"github.com/kriku/kpukbot/internal/config"
	"github.com/kriku/kpukbot/internal/handlers"
	"github.com/kriku/kpukbot/internal/logger"
	repositories "github.com/kriku/kpukbot/internal/repository/messages"
	services "github.com/kriku/kpukbot/internal/services/messages"
)

var baseSet = wire.NewSet(
	config.NewConfig,
	logger.NewLogger,
	handlers.NewDefaultHandler,
	clients.NewTelegramClient,
	services.NewTelegramMessagesService,
	repositories.NewFirestoreRepository,

	NewApp,
)

func InitAppLocal() (App, error) {
	wire.Build(
		baseSet,
	)
	return App{}, nil
}
