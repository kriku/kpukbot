//go:build wireinject
// +build wireinject

package app

import (
	"context"
	"log/slog"

	"cloud.google.com/go/firestore"
	"github.com/go-telegram/bot"
	"github.com/google/wire"

	"github.com/kriku/kpukbot/internal/clients/gemini"
	clients "github.com/kriku/kpukbot/internal/clients/telegram"
	"github.com/kriku/kpukbot/internal/config"
	"github.com/kriku/kpukbot/internal/handlers"
	"github.com/kriku/kpukbot/internal/logger"
	messagesRepo "github.com/kriku/kpukbot/internal/repository/messages"
	threadsRepo "github.com/kriku/kpukbot/internal/repository/threads"
	"github.com/kriku/kpukbot/internal/services/orchestrator"
	"github.com/kriku/kpukbot/internal/services/response"
	"github.com/kriku/kpukbot/internal/services/threading"
	"github.com/kriku/kpukbot/internal/strategies"
)

// Provider functions

// ProvideGeminiClient provides a Gemini client
func ProvideGeminiClient(ctx context.Context, cfg *config.Config, logger *slog.Logger) (gemini.Client, error) {
	return gemini.NewGeminiClient(ctx, cfg.GeminiAPIKey, logger)
}

// ProvideMessagesRepository provides a messages repository
func ProvideMessagesRepository(client *firestore.Client) messagesRepo.MessagesRepository {
	return messagesRepo.NewFirestoreMessagesRepository(client)
}

// ProvideThreadsRepository provides a threads repository
func ProvideThreadsRepository(client *firestore.Client) threadsRepo.ThreadsRepository {
	return threadsRepo.NewFirestoreThreadsRepository(client)
}

// ProvideStrategies provides all response strategies
func ProvideStrategies(geminiClient gemini.Client, logger *slog.Logger) []strategies.ResponseStrategy {
	return []strategies.ResponseStrategy{
		strategies.NewFactCheckerStrategy(geminiClient, logger),
		strategies.NewAgreementStrategy(geminiClient, logger),
		strategies.NewReminderStrategy(geminiClient, logger),
		strategies.NewGeneralStrategy(geminiClient, logger),
	}
}

// ProvideClassifierService provides the classifier service
func ProvideClassifierService(
	geminiClient gemini.Client,
	threadsRepository threadsRepo.ThreadsRepository,
	messagesRepository messagesRepo.MessagesRepository,
	logger *slog.Logger,
) *threading.ClassifierService {
	return threading.NewClassifierService(geminiClient, threadsRepository, messagesRepository, logger)
}

// ProvideAnalyzerService provides the analyzer service
func ProvideAnalyzerService(
	geminiClient gemini.Client,
	strategies []strategies.ResponseStrategy,
	logger *slog.Logger,
) *response.AnalyzerService {
	return response.NewAnalyzerService(geminiClient, strategies, logger)
}

// ProvideOrchestratorService provides the orchestrator service
func ProvideOrchestratorService(
	classifier *threading.ClassifierService,
	analyzer *response.AnalyzerService,
	messagesRepository messagesRepo.MessagesRepository,
	logger *slog.Logger,
) *orchestrator.OrchestratorService {
	// Note: TelegramClient will be set later in NewApp to avoid circular dependency
	return orchestrator.NewOrchestratorService(classifier, analyzer, messagesRepository, nil, logger)
}

// ProvideOrchestratorHandler provides the orchestrator handler
func ProvideOrchestratorHandler(
	orch *orchestrator.OrchestratorService,
	logger *slog.Logger,
) bot.HandlerFunc {
	return handlers.NewOrchestratorHandler(orch, logger)
}

var baseSet = wire.NewSet(
	// Config and context
	config.NewConfig,

	// Logger
	logger.NewLogger,

	// Firestore client
	NewFirestoreClient,

	// Gemini client
	ProvideGeminiClient,

	// Repositories
	ProvideMessagesRepository,
	ProvideThreadsRepository,

	// Services
	ProvideStrategies,
	ProvideClassifierService,
	ProvideAnalyzerService,
	ProvideOrchestratorService,

	// Handler
	ProvideOrchestratorHandler,

	// Telegram client
	clients.NewTelegramClient,

	// App
	NewApp,
)

func InitApp(ctx context.Context) (App, error) {
	wire.Build(
		baseSet,
	)
	return App{}, nil
}
