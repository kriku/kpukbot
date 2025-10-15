package orchestrator

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kriku/kpukbot/internal/clients/telegram"
	"github.com/kriku/kpukbot/internal/models"
	messagesRepo "github.com/kriku/kpukbot/internal/repository/messages"
	"github.com/kriku/kpukbot/internal/services/response"
	"github.com/kriku/kpukbot/internal/services/threading"
	"github.com/kriku/kpukbot/internal/services/users"
)

// OrchestratorService coordinates the entire message processing pipeline
type OrchestratorService struct {
	classifier     *threading.ClassifierService
	analyzer       *response.AnalyzerService
	messagesRepo   messagesRepo.MessagesRepository
	usersService   *users.UsersService
	telegramClient telegram.MessengerClient
	logger         *slog.Logger
}

func NewOrchestratorService(
	classifier *threading.ClassifierService,
	analyzer *response.AnalyzerService,
	messagesRepo messagesRepo.MessagesRepository,
	usersService *users.UsersService,
	telegramClient telegram.MessengerClient,
	logger *slog.Logger,
) *OrchestratorService {
	return &OrchestratorService{
		classifier:     classifier,
		analyzer:       analyzer,
		messagesRepo:   messagesRepo,
		usersService:   usersService,
		telegramClient: telegramClient,
		logger:         logger.With("service", "orchestrator"),
	}
}

// SetTelegramClient sets the telegram client (useful for resolving circular dependencies)
func (s *OrchestratorService) SetTelegramClient(client telegram.MessengerClient) {
	s.telegramClient = client
}

// ProcessMessage orchestrates the entire message processing flow
func (s *OrchestratorService) ProcessMessage(ctx context.Context, message *models.Message) error {
	s.logger.InfoContext(ctx, "Processing message",
		"message_id", message.ID,
		"chat_id", message.ChatID,
		"user_id", message.UserID)

	// Step 1: Save the incoming message
	if err := s.messagesRepo.SaveMessage(ctx, *message); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}
	s.logger.DebugContext(ctx, "Message saved")

	// Step 1.5: Track/update user information
	if err := s.trackUserFromMessage(ctx, message); err != nil {
		s.logger.WarnContext(ctx, "Failed to track user", "user_id", message.UserID, "error", err)
		// Don't fail the entire process if user tracking fails
	}

	// Step 2: Classify message into a thread
	threadMatch, err := s.classifier.ClassifyMessage(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to classify message: %w", err)
	}

	s.logger.InfoContext(ctx, "Message classified",
		"thread_id", threadMatch.Thread.ID,
		"thread_theme", threadMatch.Thread.Theme,
		"probability", threadMatch.Probability)

	// Step 3: Add message to the thread
	if err := s.classifier.AddMessageToThread(ctx, threadMatch.Thread, message); err != nil {
		return fmt.Errorf("failed to add message to thread: %w", err)
	}
	s.logger.DebugContext(ctx, "Message added to thread")

	// Step 4: Get recent messages from the thread for context
	messages, err := s.getThreadMessages(ctx, threadMatch.Thread)
	if err != nil {
		s.logger.WarnContext(ctx, "Failed to get thread messages", "error", err)
		messages = []*models.Message{message}
	}

	// Step 5: Analyze if response is needed and generate it
	responseText, err := s.analyzer.AnalyzeAndRespond(ctx, threadMatch.Thread, messages, message)
	if err != nil {
		return fmt.Errorf("failed to analyze and respond: %w", err)
	}

	// Step 6: Send response if generated
	if responseText != "" {
		s.logger.InfoContext(ctx, "Sending response", "response_length", len(responseText))

		_, err := s.telegramClient.SendMessage(ctx, message.ChatID, responseText)
		if err != nil {
			return fmt.Errorf("failed to send response: %w", err)
		}

		s.logger.InfoContext(ctx, "Response sent successfully")
	} else {
		s.logger.InfoContext(ctx, "No response needed")
	}

	return nil
}

// getThreadMessages retrieves recent messages from a thread
func (s *OrchestratorService) getThreadMessages(ctx context.Context, thread *models.Thread) ([]*models.Message, error) {
	var messages []*models.Message

	// Get last 10 messages from the thread
	startIdx := len(thread.MessageIDs) - 10
	if startIdx < 0 {
		startIdx = 0
	}

	for i := startIdx; i < len(thread.MessageIDs); i++ {
		msgList, err := s.messagesRepo.GetMessage(ctx, int64(thread.MessageIDs[i]))
		if err != nil {
			s.logger.WarnContext(ctx, "Failed to get message", "message_id", thread.MessageIDs[i], "error", err)
			continue
		}
		messages = append(messages, msgList...)
	}

	return messages, nil
}

// trackUserFromMessage extracts user information from the message and updates the user repository
func (s *OrchestratorService) trackUserFromMessage(ctx context.Context, message *models.Message) error {
	// Create or update user based on message information
	err := s.usersService.CreateOrUpdateUser(
		ctx,
		message.UserID,
		message.ChatID,
		message.FirstName,
		message.LastName,
		message.Username,
	)
	if err != nil {
		return fmt.Errorf("failed to create or update user: %w", err)
	}

	s.logger.DebugContext(ctx, "User information tracked",
		"user_id", message.UserID,
		"username", message.Username)

	return nil
}
