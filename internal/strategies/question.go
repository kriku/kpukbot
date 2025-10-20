package strategies

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kriku/kpukbot/internal/clients/gemini"
	"github.com/kriku/kpukbot/internal/models"
	"github.com/kriku/kpukbot/internal/prompts"
	"github.com/kriku/kpukbot/internal/services/chats"
	"github.com/kriku/kpukbot/internal/services/messages"
	"github.com/kriku/kpukbot/internal/services/users"
	"google.golang.org/genai"
)

type QuestionStrategy struct {
	gemini         gemini.Client
	userService    *users.UsersService
	chatService    *chats.ChatsService
	messageService *messages.TelegramMessagesService
	logger         *slog.Logger
}

func NewQuestionStrategy(
	gemini gemini.Client,
	userService *users.UsersService,
	chatService *chats.ChatsService,
	messageService *messages.TelegramMessagesService,
	logger *slog.Logger,
) *QuestionStrategy {
	return &QuestionStrategy{
		gemini:         gemini,
		userService:    userService,
		chatService:    chatService,
		messageService: messageService,
		logger:         logger.With("strategy", "question"),
	}
}

func (s *QuestionStrategy) Name() string {
	return "question"
}

func (s *QuestionStrategy) Priority() int {
	return 90 // High priority for triggered questions
}

func (s *QuestionStrategy) ShouldRespond(ctx context.Context, thread *models.Thread, messages []*models.Message, newMessage *models.Message) (bool, float64, error) {
	// This strategy is triggered programmatically, not by user messages
	// It will only respond when explicitly invoked via the question trigger
	return false, 0.0, nil
}

func (s *QuestionStrategy) GenerateResponse(ctx context.Context, thread *models.Thread, messages []*models.Message, newMessage *models.Message) (string, error) {
	// This method should not be called directly for question strategy
	// Questions are generated via the AskQuestionToUser method
	return "", nil
}

// AskQuestionToUser generates and asks a question to the next user in queue for the given chat
func (s *QuestionStrategy) AskQuestionToUser(ctx context.Context, chatID int64) (string, int64, error) {
	// Get the next user in queue
	queueEntry, err := s.chatService.GetNextUserInQueue(ctx, chatID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get next user in queue", "chat_id", chatID, "error", err)
		return "", 0, err
	}

	if queueEntry == nil {
		s.logger.InfoContext(ctx, "No users in queue", "chat_id", chatID)
		return "No users are currently in the question queue.", 0, nil
	}

	// Get user details to understand their interests and hobbies
	user, err := s.userService.GetUser(ctx, queueEntry.UserID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get user details", "user_id", queueEntry.UserID, "error", err)
		return "", 0, err
	}

	if user == nil {
		s.logger.WarnContext(ctx, "User not found", "user_id", queueEntry.UserID)
		return "", 0, nil
	}

	// Generate a question based on user's interests and hobbies
	question, err := s.generateQuestionForUser(ctx, user)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to generate question", "user_id", user.ID, "error", err)
		return "", 0, err
	}

	return question, queueEntry.UserID, nil
}

// generateQuestionForUser creates a personalized question based on user's interests and hobbies
func (s *QuestionStrategy) generateQuestionForUser(ctx context.Context, user *models.User) (string, error) {
	prompt := prompts.QuestionGenerationPrompt(user)

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText("Generate an engaging, thoughtful question based on the user's interests and hobbies. Keep it conversational and interesting. NEVER repeat or rephrase previously asked questions - create completely fresh, unique questions that explore different aspects. Maximum 300 characters.", genai.RoleModel),
		ResponseMIMEType:  "text/plain",
	}

	response, err := s.gemini.GenerateContent(ctx, prompt, config)
	if err != nil {
		return "", err
	}

	s.logger.InfoContext(ctx, "Generated question for user",
		"user_id", user.ID,
		"question_length", len(response),
		"interests_count", len(user.Interests),
		"hobbies_count", len(user.Hobbies))

	return response, nil
}

// SaveQuestionAsMessage saves a question as a bot message to maintain conversation history
func (s *QuestionStrategy) SaveQuestionAsMessage(ctx context.Context, chatID int64, messageID int, questionText string) error {
	return s.messageService.SaveBotMessage(ctx, chatID, messageID, questionText)
}

// RephraseQuestionForUser rephrases an existing question for a user to make it more engaging
func (s *QuestionStrategy) RephraseQuestionForUser(ctx context.Context, user *models.User, originalQuestion string) (string, error) {
	prompt := prompts.QuestionRephrasePrompt(user, originalQuestion)

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText("Rephrase the existing question to be more engaging and include a user mention. Keep the core intent but make it sound fresh and encouraging. Maximum 400 characters.", genai.RoleModel),
		ResponseMIMEType:  "text/plain",
	}

	response, err := s.gemini.GenerateContent(ctx, prompt, config)
	if err != nil {
		return "", err
	}

	s.logger.InfoContext(ctx, "Generated rephrased question for user",
		"user_id", user.ID,
		"original_length", len(originalQuestion),
		"rephrased_length", len(response),
		"interests_count", len(user.Interests),
		"hobbies_count", len(user.Hobbies))

	return response, nil
}

// MarkQuestionAsAsked marks a question as asked using the actual message ID
func (s *QuestionStrategy) MarkQuestionAsAsked(ctx context.Context, chatID int64, userID int64, messageID int) error {
	// Add message ID to the queue entry
	questionID := int64(messageID)
	err := s.chatService.AddMessageToQueueEntry(ctx, chatID, userID, questionID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to add question message to queue entry", "user_id", userID, "message_id", messageID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "Question message added to queue entry", "user_id", userID, "chat_id", chatID, "message_id", messageID)
	return nil
}

// MarkQuestionAsAskedWithText marks a question as asked and saves it to user's previous questions
func (s *QuestionStrategy) MarkQuestionAsAskedWithText(ctx context.Context, chatID int64, userID int64, messageID int, questionText string) error {
	// First mark the question as asked in the chat queue
	err := s.MarkQuestionAsAsked(ctx, chatID, userID, messageID)
	if err != nil {
		return err
	}

	// Then save the question text to the user's previous questions
	err = s.userService.AddPreviousQuestion(ctx, userID, questionText, int64(messageID))
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to save question to user history", "user_id", userID, "message_id", messageID, "error", err)
		// Don't return error - the question was already marked as asked in the queue
		// Just log the issue
	} else {
		s.logger.InfoContext(ctx, "Question saved to user history", "user_id", userID, "message_id", messageID)
	}

	return nil
}

// GetQuestionByID retrieves the original question text using the stored question ID
func (s *QuestionStrategy) GetQuestionByID(ctx context.Context, questionID int64) (string, error) {
	// Question ID is directly the message ID
	messageID := int(questionID)

	// Get the message from the database
	message, err := s.messageService.GetMessageByID(ctx, messageID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get message by ID", "message_id", messageID, "error", err)
		return "", fmt.Errorf("failed to get message: %w", err)
	}

	s.logger.InfoContext(ctx, "Retrieved question by ID",
		"question_id", questionID,
		"message_id", messageID,
		"text_length", len(message.Text))

	return message.Text, nil
}

// GetUser exposes the user service method for external access
func (s *QuestionStrategy) GetUser(ctx context.Context, userID int64) (*models.User, error) {
	return s.userService.GetUser(ctx, userID)
}
