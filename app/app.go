package app

import (
	"context"
	"log/slog"

	"cloud.google.com/go/firestore"
	clients "github.com/kriku/kpukbot/internal/clients/telegram"
	repositories "github.com/kriku/kpukbot/internal/repository/messages"
	"github.com/kriku/kpukbot/internal/services/orchestrator"
)

type App struct {
	Context            context.Context
	Logger             *slog.Logger
	MessengerClient    clients.MessengerClient
	MessagesRepository repositories.MessagesRepository
	Orchestrator       *orchestrator.OrchestratorService
	FirestoreClient    *firestore.Client
}

func NewApp(
	ctx context.Context,
	lo *slog.Logger,
	mc clients.MessengerClient,
	mr repositories.MessagesRepository,
	orch *orchestrator.OrchestratorService,
	fc *firestore.Client,
) App {
	// Set the telegram client in the orchestrator to resolve circular dependency
	orch.SetTelegramClient(mc)

	return App{
		Context:            ctx,
		Logger:             lo,
		MessengerClient:    mc,
		MessagesRepository: mr,
		Orchestrator:       orch,
		FirestoreClient:    fc,
	}
}

// Close closes all resources
func (a *App) Close() error {
	if a.FirestoreClient != nil {
		return a.FirestoreClient.Close()
	}
	return nil
}
