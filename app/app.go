package app

import (
	"log/slog"

	"cloud.google.com/go/firestore"
	clients "github.com/kriku/kpukbot/internal/clients/telegram"
	repositories "github.com/kriku/kpukbot/internal/repository/messages"
	"github.com/kriku/kpukbot/internal/services/chats"
	"github.com/kriku/kpukbot/internal/services/orchestrator"
	"github.com/kriku/kpukbot/internal/strategies"
)

type App struct {
	Logger             *slog.Logger
	MessengerClient    clients.MessengerClient
	MessagesRepository repositories.MessagesRepository
	Orchestrator       *orchestrator.OrchestratorService
	FirestoreClient    *firestore.Client
	ChatsService       *chats.ChatsService
	Strategies         []strategies.ResponseStrategy
}

func NewApp(
	lo *slog.Logger,
	mc clients.MessengerClient,
	mr repositories.MessagesRepository,
	orch *orchestrator.OrchestratorService,
	fc *firestore.Client,
	cs *chats.ChatsService,
	strats []strategies.ResponseStrategy,
) App {
	// Set the telegram client in the orchestrator to resolve circular dependency
	orch.SetTelegramClient(mc)

	return App{
		Logger:             lo,
		MessengerClient:    mc,
		MessagesRepository: mr,
		Orchestrator:       orch,
		FirestoreClient:    fc,
		ChatsService:       cs,
		Strategies:         strats,
	}
}

// Close closes all resources
func (a *App) Close() error {
	if a.FirestoreClient != nil {
		return a.FirestoreClient.Close()
	}
	return nil
}
