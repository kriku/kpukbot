package app

import (
	"github.com/kriku/kpukbot/internal/handlers"
	"github.com/kriku/kpukbot/internal/services"
)

type App struct {
	TelegramService *services.TelegramService
	AIService       services.AIServiceInterface
	WebhookHandler  *handlers.WebhookHandler
}

func NewApp(
	tg *services.TelegramService,
	ai services.AIServiceInterface,
	wh *handlers.WebhookHandler,
) App {
	return App{
		TelegramService: tg,
		AIService:       ai,
		WebhookHandler:  wh,
	}
}
