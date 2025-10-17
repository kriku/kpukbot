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
	"github.com/kriku/kpukbot/internal/services/messages"
	"github.com/kriku/kpukbot/internal/services/users"
	"google.golang.org/genai"
)

type AssessmentStrategy struct {
	gemini         gemini.Client
	userService    *users.UsersService
	messageService *messages.TelegramMessagesService
	logger         *slog.Logger
}

func NewAssessmentStrategy(
	gemini gemini.Client,
	userService *users.UsersService,
	messageService *messages.TelegramMessagesService,
	logger *slog.Logger,
) *AssessmentStrategy {
	return &AssessmentStrategy{
		gemini:         gemini,
		userService:    userService,
		messageService: messageService,
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

	// Use LLM to determine if this message should trigger assessment
	shouldRespond, confidence, err := s.evaluateShouldRespondWithLLM(ctx, thread, messages, newMessage, user)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to evaluate with LLM", "error", err)
		return false, 0.0, err
	}

	s.logger.InfoContext(ctx, "LLM assessment strategy evaluation complete",
		"should_respond", shouldRespond,
		"confidence", confidence,
		"user_id", newMessage.UserID)

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

	// Generate assessment using multi-turn conversation
	assessment, err := s.assessUserAnswer(ctx, user, conversationHistory)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to generate assessment", "error", err)
		return "", fmt.Errorf("failed to assess answer: %w", err)
	}

	s.logger.InfoContext(ctx, "Generated assessment",
		"user_id", newMessage.UserID,
		"assessment_score", assessment.Score,
		"assessment_feedback", assessment.Feedback)

	// Format the response
	response := s.formatAssessmentResponse(assessment)

	return response, nil
}

// evaluateShouldRespondWithLLM uses LLM to determine if this message should trigger assessment
func (s *AssessmentStrategy) evaluateShouldRespondWithLLM(ctx context.Context, thread *models.Thread, messages []*models.Message, newMessage *models.Message, user *models.User) (bool, float64, error) {
	// Build conversation context
	conversationContext := s.buildConversationContext(messages, newMessage)

	// Create evaluation prompt using centralized template
	prompt := prompts.AssessmentShouldRespondPrompt(thread, conversationContext, newMessage, user)

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
			Required: []string{"should_respond", "confidence", "reason"},
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
func (s *AssessmentStrategy) assessUserAnswer(ctx context.Context, user *models.User, conversationHistory []gemini.Message) (*AssessmentResult, error) {
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

	// Build the full prompt with conversation context and assessment instructions
	fullPrompt := fmt.Sprintf(
		`%s\n\nConversation History: %s\n\nPlease assess the user's latest response in the conversation context.`,
		assessmentPrompt,
		conversationContext.String(),
	)

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText("Assess the user's response quality and provide constructive feedback. Return valid JSON with score, feedback, and optional follow-up question.", genai.RoleModel),
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
					MaxLength: &constants.MaxAnalysisLength,
				},
			},
			Required: []string{"score", "feedback", "follow_up_needed"},
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

	// Add follow-up question if needed
	if assessment.FollowUpNeeded && assessment.FollowUpQuestion != "" {
		response += "\n\n" + assessment.FollowUpQuestion
	}

	return response
}
