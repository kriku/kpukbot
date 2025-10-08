package factory

import (
	"context"
	"fmt"
	"log/slog"

	"cloud.google.com/go/firestore"
	"github.com/kriku/kpukbot/internal/clients/gemini"
	"github.com/kriku/kpukbot/internal/config"
	messagesRepo "github.com/kriku/kpukbot/internal/repository/messages"
	threadsRepo "github.com/kriku/kpukbot/internal/repository/threads"
	"github.com/kriku/kpukbot/internal/services/orchestrator"
	"github.com/kriku/kpukbot/internal/services/response"
	"github.com/kriku/kpukbot/internal/services/threading"
	"github.com/kriku/kpukbot/internal/strategies"
)

// ServiceFactory creates and wires all services together
type ServiceFactory struct {
	config          *config.Config
	logger          *slog.Logger
	firestoreClient *firestore.Client
	geminiClient    gemini.Client
}

func NewServiceFactory(cfg *config.Config, logger *slog.Logger, firestoreClient *firestore.Client) *ServiceFactory {
	return &ServiceFactory{
		config:          cfg,
		logger:          logger,
		firestoreClient: firestoreClient,
	}
}

// CreateOrchestrator creates the main orchestrator with all dependencies
func (f *ServiceFactory) CreateOrchestrator(ctx context.Context) (*orchestrator.OrchestratorService, error) {
	// Create Gemini client
	geminiClient, err := f.GetGeminiClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Create repositories
	messagesRepository := messagesRepo.NewFirestoreMessagesRepository(f.firestoreClient)
	threadsRepository := threadsRepo.NewFirestoreThreadsRepository(f.firestoreClient)

	// Create classifier service
	classifierService := threading.NewClassifierService(
		geminiClient,
		threadsRepository,
		messagesRepository,
		f.logger,
	)

	// Create strategies
	strategies := f.CreateStrategies(geminiClient)

	// Create analyzer service
	analyzerService := response.NewAnalyzerService(
		geminiClient,
		strategies,
		f.logger,
	)

	// Note: telegram client should be passed in from outside
	// since it needs the orchestrator (circular dependency)
	// We'll set it later
	orchestratorService := orchestrator.NewOrchestratorService(
		classifierService,
		analyzerService,
		messagesRepository,
		nil, // Will be set later
		f.logger,
	)

	return orchestratorService, nil
}

// GetGeminiClient returns or creates the Gemini client
func (f *ServiceFactory) GetGeminiClient(ctx context.Context) (gemini.Client, error) {
	if f.geminiClient != nil {
		return f.geminiClient, nil
	}

	client, err := gemini.NewGeminiClient(ctx, f.config.GeminiAPIKey, f.logger)
	if err != nil {
		return nil, err
	}

	f.geminiClient = client
	return client, nil
}

// CreateStrategies creates all response strategies
func (f *ServiceFactory) CreateStrategies(geminiClient gemini.Client) []strategies.ResponseStrategy {
	return []strategies.ResponseStrategy{
		strategies.NewFactCheckerStrategy(geminiClient, f.logger),
		strategies.NewReminderStrategy(geminiClient, f.logger),
		strategies.NewAgreementStrategy(geminiClient, f.logger),
		strategies.NewGeneralStrategy(geminiClient, f.logger),
	}
}

// Close closes all resources
func (f *ServiceFactory) Close() error {
	if f.geminiClient != nil {
		if err := f.geminiClient.Close(); err != nil {
			return err
		}
	}
	return nil
}
