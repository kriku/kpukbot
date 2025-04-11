//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"

	"github.com/kriku/kpukbot/internal/config"
	"github.com/kriku/kpukbot/internal/handlers"
	"github.com/kriku/kpukbot/internal/services"
)

var baseSet = wire.NewSet(
	config.NewConfig,
	handlers.NewWebhookHandler,
	services.NewGeminiService,
	services.NewTelegramService,
	wire.Bind(new(handlers.TelegramUpdateHandler), new(*services.TelegramService)),

	NewApp,
)

func InitAppLocal() (App, error) {
	wire.Build(
		baseSet,
	)
	return App{}, nil
}
