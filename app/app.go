package app

import (
	"github.com/kriku/kpukbot/internal/handlers"
	"github.com/kriku/kpukbot/internal/repository"
	"github.com/kriku/kpukbot/internal/services"
)

type App struct {
	TelegramService *services.TelegramService
	AIService       services.AIServiceInterface
	WebhookHandler  *handlers.WebhookHandler
	Repository      repository.Repository
}

func NewApp(
	tg *services.TelegramService,
	ai services.AIServiceInterface,
	wh *handlers.WebhookHandler,
	repo repository.Repository,
) App {
	return App{
		TelegramService: tg,
		AIService:       ai,
		WebhookHandler:  wh,
		Repository:      repo,
	}
}
