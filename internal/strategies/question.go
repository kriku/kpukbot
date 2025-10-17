package strategies

import (
	"context"
	"log/slog"

	"github.com/kriku/kpukbot/internal/clients/gemini"
	"github.com/kriku/kpukbot/internal/models"
	"github.com/kriku/kpukbot/internal/prompts"
	"github.com/kriku/kpukbot/internal/services/chats"
	"github.com/kriku/kpukbot/internal/services/users"
	"google.golang.org/genai"
)

type QuestionStrategy struct {
	gemini      gemini.Client
	userService *users.UsersService
	chatService *chats.ChatsService
	logger      *slog.Logger
}

func NewQuestionStrategy(gemini gemini.Client, userService *users.UsersService, chatService *chats.ChatsService, logger *slog.Logger) *QuestionStrategy {
	return &QuestionStrategy{
		gemini:      gemini,
		userService: userService,
		chatService: chatService,
		logger:      logger.With("strategy", "question"),
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

	// Mark the question as asked for tracking
	questionID := "q_" + string(rune(queueEntry.UserID)) + "_" + string(rune(chatID))
	err = s.chatService.MarkQuestionAsked(ctx, chatID, queueEntry.UserID, questionID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to mark question as asked", "user_id", queueEntry.UserID, "error", err)
		// Don't return error - question can still be asked
	}

	return question, queueEntry.UserID, nil
}

// generateQuestionForUser creates a personalized question based on user's interests and hobbies
func (s *QuestionStrategy) generateQuestionForUser(ctx context.Context, user *models.User) (string, error) {
	prompt := prompts.QuestionGenerationPrompt(user)

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText("Generate an engaging, thoughtful question based on the user's interests and hobbies. Keep it conversational and interesting. Maximum 300 characters.", genai.RoleModel),
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
