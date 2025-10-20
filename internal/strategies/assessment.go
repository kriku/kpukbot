package strategies

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/kriku/kpukbot/internal/clients/gemini"
	"github.com/kriku/kpukbot/internal/constants"
	"github.com/kriku/kpukbot/internal/models"
	"github.com/kriku/kpukbot/internal/prompts"
	messagesRepo "github.com/kriku/kpukbot/internal/repository/messages"
	"github.com/kriku/kpukbot/internal/services/chats"
	"github.com/kriku/kpukbot/internal/services/messages"
	"github.com/kriku/kpukbot/internal/services/users"
	"google.golang.org/genai"
)

// No longer using string prefixes - we rely on database tracking

type AssessmentStrategy struct {
	gemini         gemini.Client
	userService    *users.UsersService
	messageService *messages.TelegramMessagesService
	chatsService   *chats.ChatsService
	messagesRepo   messagesRepo.MessagesRepository
	logger         *slog.Logger
}

func NewAssessmentStrategy(
	gemini gemini.Client,
	userService *users.UsersService,
	messageService *messages.TelegramMessagesService,
	chatsService *chats.ChatsService,
	messagesRepo messagesRepo.MessagesRepository,
	logger *slog.Logger,
) *AssessmentStrategy {
	return &AssessmentStrategy{
		gemini:         gemini,
		userService:    userService,
		messageService: messageService,
		chatsService:   chatsService,
		messagesRepo:   messagesRepo,
		logger:         logger.With("strategy", "assessment"),
	}
}

func (s *AssessmentStrategy) Name() string {
	return "assessment"
}

func (s *AssessmentStrategy) Priority() int {
	return 85 // High priority for answer assessment
}

func (s *AssessmentStrategy) ShouldRespond(ctx context.Context, thread *models.Thread, messages []*models.Message, newMessage *models.Message) (bool, float64, error) {
	s.logger.InfoContext(ctx, "Evaluating if assessment strategy should respond using LLM",
		"thread_id", thread.ID,
		"user_id", newMessage.UserID,
		"message_text", newMessage.Text)

	// Get user context if available (skip if userService is nil for testing)
	var user *models.User
	if s.userService != nil {
		var err error
		user, err = s.userService.GetUser(ctx, newMessage.UserID)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to get user", "error", err, "user_id", newMessage.UserID)
			return false, 0.0, err
		}

		if user == nil {
			s.logger.WarnContext(ctx, "User not found", "user_id", newMessage.UserID)
			return false, 0.0, nil
		}
	}

	// Check if user is currently being asked a question in the chat queue
	var isBeingAsked bool
	if s.chatsService != nil && thread != nil {
		chat, err := s.chatsService.GetChat(ctx, thread.ChatID)
		if err == nil && chat != nil {
			for _, entry := range chat.QuestionQueue {
				if entry.UserID == newMessage.UserID && entry.Status == models.QueueStatusAsking {
					isBeingAsked = true
					s.logger.InfoContext(ctx, "User is currently being asked a question",
						"user_id", newMessage.UserID,
						"question_ids", entry.QuestionIDs)
					break
				}
			}
		}
	}

	// Use LLM to determine if this message should trigger assessment
	shouldRespond, confidence, err := s.evaluateShouldRespondWithLLM(ctx, thread, messages, newMessage, user, isBeingAsked)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to evaluate with LLM", "error", err)
		return false, 0.0, err
	}

	s.logger.InfoContext(ctx, "LLM assessment strategy evaluation complete",
		"should_respond", shouldRespond,
		"confidence", confidence,
		"user_id", newMessage.UserID,
		"is_being_asked", isBeingAsked)

	return shouldRespond, confidence, nil
}

func (s *AssessmentStrategy) GenerateResponse(ctx context.Context, thread *models.Thread, messages []*models.Message, newMessage *models.Message) (string, error) {
	s.logger.InfoContext(ctx, "Generating assessment response",
		"thread_id", thread.ID,
		"user_id", newMessage.UserID)

	// Get conversation history for multi-turn context
	conversationHistory := s.buildConversationHistory(messages, newMessage)

	// Get user details for context
	user, err := s.userService.GetUser(ctx, newMessage.UserID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get user for assessment", "error", err)
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Check if this message answers a previous follow-up question
	unansweredFollowUp, err := s.getUnansweredFollowUpQuestion(ctx, thread.ChatID, newMessage.UserID)
	if err != nil {
		s.logger.WarnContext(ctx, "Failed to get unanswered follow-up questions", "error", err)
	} else if unansweredFollowUp != nil {
		s.logger.InfoContext(ctx, "User is answering a follow-up question",
			"follow_up_message_id", unansweredFollowUp.ID,
			"user_answer", newMessage.Text)
	}

	// Get recent follow-up questions for context
	recentFollowUps, err := s.getRecentFollowUpQuestions(ctx, thread.ChatID, newMessage.UserID, 5)
	if err != nil {
		s.logger.WarnContext(ctx, "Failed to get recent follow-up questions", "error", err)
		recentFollowUps = []*models.Message{} // Continue without follow-up context
	}

	// Generate assessment using multi-turn conversation and follow-up context
	assessment, err := s.assessUserAnswer(ctx, user, conversationHistory, recentFollowUps)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to generate assessment", "error", err)
		return "", fmt.Errorf("failed to assess answer: %w", err)
	}

	s.logger.InfoContext(ctx, "Generated assessment",
		"user_id", newMessage.UserID,
		"assessment_score", assessment.Score,
		"assessment_feedback", assessment.Feedback,
		"follow_up_needed", assessment.FollowUpNeeded)

	// Store the user's answer message ID in the queue entry
	if s.chatsService != nil && thread != nil {
		err := s.chatsService.AddMessageToQueueEntry(ctx, thread.ChatID, newMessage.UserID, int64(newMessage.ID))
		if err != nil {
			s.logger.WarnContext(ctx, "Failed to add user answer to queue entry",
				"error", err,
				"user_id", newMessage.UserID,
				"message_id", newMessage.ID)
			// Don't fail the assessment - this is just for tracking
		} else {
			s.logger.InfoContext(ctx, "User answer added to queue entry",
				"user_id", newMessage.UserID,
				"message_id", newMessage.ID)
		}
	}

	// Note: Follow-up question storage will be handled by the orchestrator
	// when it processes the response with the follow-up delimiter.
	// The follow-up question will be automatically stored as a bot message.

	// Mark question as completed if assessment criteria are met
	if s.chatsService != nil && thread != nil {
		// Mark as completed if score is decent (>= 0.6) and no follow-up is needed
		if assessment.Score >= 0.6 && !assessment.FollowUpNeeded {
			err := s.chatsService.MarkQuestionAnsweredAndReEnqueue(ctx, thread.ChatID, newMessage.UserID)
			if err != nil {
				s.logger.ErrorContext(ctx, "Failed to mark question as completed and re-enqueue",
					"error", err,
					"chat_id", thread.ChatID,
					"user_id", newMessage.UserID)
				// Don't return error - assessment response should still be sent
			} else {
				s.logger.InfoContext(ctx, "Question marked as completed and user re-enqueued",
					"chat_id", thread.ChatID,
					"user_id", newMessage.UserID,
					"score", assessment.Score)
			}
		} else {
			s.logger.InfoContext(ctx, "Question not marked as completed",
				"chat_id", thread.ChatID,
				"user_id", newMessage.UserID,
				"score", assessment.Score,
				"follow_up_needed", assessment.FollowUpNeeded,
				"reason", "Score too low or follow-up needed")
		}
	}

	// Format the response (including follow-up question if needed)
	response := s.formatAssessmentResponse(assessment)

	return response, nil
}

// evaluateShouldRespondWithLLM uses LLM to determine if this message should trigger assessment
func (s *AssessmentStrategy) evaluateShouldRespondWithLLM(ctx context.Context, thread *models.Thread, messages []*models.Message, newMessage *models.Message, user *models.User, isBeingAsked bool) (bool, float64, error) {
	// Build conversation context
	conversationContext := s.buildConversationContext(messages, newMessage)

	// Create evaluation prompt using centralized template
	prompt := prompts.AssessmentShouldRespondPrompt(thread, conversationContext, newMessage, user, isBeingAsked)

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText("Analyze the conversation to determine if the user's message should trigger assessment response. Be precise.", genai.RoleModel),
		ResponseMIMEType:  "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"should_respond": {
					Type: genai.TypeBoolean,
				},
				"confidence": {
					Type:    genai.TypeNumber,
					Minimum: &constants.MinimumConfidenceScore,
					Maximum: &constants.MaximumConfidenceScore,
				},
				"reason": {
					Type:      genai.TypeString,
					MaxLength: &constants.MaxAnalysisLength,
				},
			},
		},
	}

	// Use Gemini to evaluate
	response, err := s.gemini.GenerateContent(ctx, prompt, config)
	if err != nil {
		return false, 0.0, fmt.Errorf("failed to evaluate with LLM: %w", err)
	}

	// Parse the response
	var evaluation struct {
		ShouldRespond bool    `json:"should_respond"`
		Confidence    float64 `json:"confidence"`
		Reason        string  `json:"reason"`
	}

	if err := json.Unmarshal([]byte(response), &evaluation); err != nil {
		s.logger.WarnContext(ctx, "Failed to parse LLM evaluation, defaulting to false", "error", err, "response", response)
		return false, 0.0, nil
	}

	s.logger.InfoContext(ctx, "LLM evaluation result",
		"should_respond", evaluation.ShouldRespond,
		"confidence", evaluation.Confidence,
		"reason", evaluation.Reason)

	return evaluation.ShouldRespond, evaluation.Confidence, nil
}

// buildConversationContext creates a string representation of recent conversation
func (s *AssessmentStrategy) buildConversationContext(messages []*models.Message, newMessage *models.Message) string {
	var contextBuilder strings.Builder

	// Include recent messages for context (last 5-10 messages)
	recentMessages := messages
	if len(messages) > 10 {
		recentMessages = messages[len(messages)-10:]
	}

	for _, msg := range recentMessages {
		sender := "User"
		if msg.IsBot {
			sender = "Bot"
		}
		contextBuilder.WriteString(fmt.Sprintf("%s: %s\n", sender, msg.Text))
	}

	// Add the new message
	contextBuilder.WriteString(fmt.Sprintf("User: %s\n", newMessage.Text))

	return contextBuilder.String()
}

// buildConversationHistory converts messages to conversation format for Gemini
func (s *AssessmentStrategy) buildConversationHistory(messages []*models.Message, newMessage *models.Message) []gemini.Message {
	var history []gemini.Message

	// Add previous messages
	for _, msg := range messages {
		role := "user"
		if msg.IsBot {
			role = "model"
		}

		history = append(history, gemini.Message{
			Role:    role,
			Content: msg.Text,
		})
	}

	// Add the new user message
	history = append(history, gemini.Message{
		Role:    "user",
		Content: newMessage.Text,
	})

	return history
}

// AssessmentResult represents the assessment of a user's answer
type AssessmentResult struct {
	Score            float64 `json:"score"`                        // 0.0 to 1.0
	Feedback         string  `json:"feedback"`                     // Detailed feedback
	FollowUpNeeded   bool    `json:"follow_up_needed"`             // Whether a follow-up question is needed
	FollowUpQuestion string  `json:"follow_up_question,omitempty"` // Optional follow-up question
}

// assessUserAnswer uses Gemini to assess the user's answer in context
func (s *AssessmentStrategy) assessUserAnswer(ctx context.Context, user *models.User, conversationHistory []gemini.Message, recentFollowUps []*models.Message) (*AssessmentResult, error) {
	// Create the assessment prompt using centralized template
	assessmentPrompt := prompts.AssessmentEvaluationPrompt(user)

	// Build conversation context string
	var conversationContext strings.Builder
	for _, msg := range conversationHistory {
		sender := "User"
		if msg.Role == "model" {
			sender = "Bot"
		}
		conversationContext.WriteString(fmt.Sprintf("%s: %s\n", sender, msg.Content))
	}

	// Build follow-up questions context using the helper method
	followUpContext := s.buildFollowUpContext(recentFollowUps)

	// Build the full prompt with conversation context, follow-up context, and assessment instructions
	fullPrompt := fmt.Sprintf(
		`%s\n\nConversation History: %s%s\n\nPlease assess the user's latest response in the conversation context, taking into account any previous follow-up questions and answers.`,
		assessmentPrompt,
		conversationContext.String(),
		followUpContext,
	)

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText("Assess the user's response quality and provide constructive feedback. Return score, feedback, and optional follow-up question.", genai.RoleModel),
		ResponseMIMEType:  "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"score": {
					Type:    genai.TypeNumber,
					Minimum: &constants.MinimumConfidenceScore,
					Maximum: &constants.MaximumConfidenceScore,
				},
				"feedback": {
					Type:      genai.TypeString,
					MaxLength: &constants.MaxAssessmentFeedbackLength,
				},
				"follow_up_needed": {
					Type: genai.TypeBoolean,
				},
				"follow_up_question": {
					Type:      genai.TypeString,
					MaxLength: &constants.MaxAssessmentFollowUpQuestionLength,
				},
			},
		},
	}

	// Use Gemini to generate assessment
	response, err := s.gemini.GenerateContent(ctx, fullPrompt, config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate assessment: %w", err)
	}

	var assessment AssessmentResult
	if err := json.Unmarshal([]byte(response), &assessment); err != nil {
		return nil, fmt.Errorf("failed to parse assessment response: %w", err)
	}

	return &assessment, nil
}

// formatAssessmentResponse formats the assessment result into a human-readable response
func (s *AssessmentStrategy) formatAssessmentResponse(assessment *AssessmentResult) string {
	response := assessment.Feedback

	// If there's a follow-up question, include it with a special delimiter
	// This allows the orchestrator to split them into separate messages
	if assessment.FollowUpNeeded && assessment.FollowUpQuestion != "" {
		response += "\n\n---FOLLOW_UP---\n" + assessment.FollowUpQuestion
	}

	return response
}

// getRecentFollowUpQuestions retrieves recent follow-up questions for a user using queue entry data
func (s *AssessmentStrategy) getRecentFollowUpQuestions(ctx context.Context, chatID int64, userID int64, limit int) ([]*models.Message, error) {
	// Get the user's current queue entry to access question IDs
	chat, err := s.chatsService.GetChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}

	if chat == nil {
		return []*models.Message{}, nil
	}

	// Find the user's queue entry
	var userEntry *models.QueueEntry
	for _, entry := range chat.QuestionQueue {
		if entry.UserID == userID {
			userEntry = &entry
			break
		}
	}

	if userEntry == nil || len(userEntry.QuestionIDs) == 0 {
		return []*models.Message{}, nil
	}

	// Get all messages in the conversation session
	allMessages, err := s.messagesRepo.GetMessages(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// Create a map for quick message lookup by ID
	messageMap := make(map[int64]*models.Message)
	for _, msg := range allMessages {
		messageMap[int64(msg.ID)] = msg
	}

	// Extract messages from queue entry and identify follow-up questions
	var followUpQuestions []*models.Message
	var lastBotMessageID int64 = -1

	// Process messages in chronological order to identify follow-ups
	for _, msgID := range userEntry.QuestionIDs {
		if msg, exists := messageMap[msgID]; exists {
			if msg.IsBot {
				// If we've seen a bot message before, this is a follow-up question
				if lastBotMessageID != -1 {
					followUpQuestions = append(followUpQuestions, msg)
				}
				lastBotMessageID = msgID
			}
		}
	}

	// Limit the results
	if len(followUpQuestions) > limit {
		followUpQuestions = followUpQuestions[len(followUpQuestions)-limit:]
	}

	return followUpQuestions, nil
}

// getUnansweredFollowUpQuestion gets the most recent unanswered follow-up question for a user using queue data
func (s *AssessmentStrategy) getUnansweredFollowUpQuestion(ctx context.Context, chatID int64, userID int64) (*models.Message, error) {
	// Get the user's current queue entry
	chat, err := s.chatsService.GetChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}

	if chat == nil {
		return nil, nil
	}

	// Find the user's queue entry
	var userEntry *models.QueueEntry
	for _, entry := range chat.QuestionQueue {
		if entry.UserID == userID {
			userEntry = &entry
			break
		}
	}

	if userEntry == nil || len(userEntry.QuestionIDs) == 0 {
		return nil, nil
	}

	// Get all messages
	allMessages, err := s.messagesRepo.GetMessages(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// Create message map for quick lookup
	messageMap := make(map[int64]*models.Message)
	for _, msg := range allMessages {
		messageMap[int64(msg.ID)] = msg
	}

	// Find the last bot message (potential unanswered follow-up)
	// and check if there's a user response after it
	var lastBotMessage *models.Message
	var lastBotMessageIndex = -1

	// Find messages in the queue entry and identify the conversation flow
	for i, msgID := range userEntry.QuestionIDs {
		if msg, exists := messageMap[msgID]; exists {
			if msg.IsBot {
				lastBotMessage = msg
				lastBotMessageIndex = i
			} else {
				// If we find a user message after the last bot message, reset
				if lastBotMessageIndex != -1 && i > lastBotMessageIndex {
					lastBotMessage = nil
					lastBotMessageIndex = -1
				}
			}
		}
	}

	// If we have a bot message with no user response after it, it's unanswered
	// But only consider it a follow-up if it's not the first bot message in the session
	if lastBotMessage != nil {
		// Check if this is a follow-up (not the first bot message)
		botMessageCount := 0
		for _, msgID := range userEntry.QuestionIDs {
			if msg, exists := messageMap[msgID]; exists && msg.IsBot {
				botMessageCount++
				if msg.ID == lastBotMessage.ID && botMessageCount > 1 {
					return lastBotMessage, nil
				}
			}
		}
	}

	return nil, nil
}

// buildFollowUpContext creates context string from follow-up messages
func (s *AssessmentStrategy) buildFollowUpContext(followUpMessages []*models.Message) string {
	if len(followUpMessages) == 0 {
		return ""
	}

	var context strings.Builder
	context.WriteString("\n\nPrevious Follow-up Questions Context:\n")

	for i, msg := range followUpMessages {
		context.WriteString(fmt.Sprintf("Follow-up %d: %s\n", i+1, msg.Text))
	}

	return context.String()
}
