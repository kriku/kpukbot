package strategies

import (
	"context"
	"log/slog"

	"github.com/kriku/kpukbot/internal/clients/gemini"
	"github.com/kriku/kpukbot/internal/constants"
	"github.com/kriku/kpukbot/internal/models"
	"github.com/kriku/kpukbot/internal/prompts"
	"github.com/kriku/kpukbot/internal/services/users"
	"google.golang.org/genai"
)

type IntroductionStrategy struct {
	gemini      gemini.Client
	userService *users.UsersService
	logger      *slog.Logger
}

func NewIntroductionStrategy(gemini gemini.Client, userService *users.UsersService, logger *slog.Logger) *IntroductionStrategy {
	return &IntroductionStrategy{
		gemini:      gemini,
		userService: userService,
		logger:      logger.With("strategy", "introduction"),
	}
}

func (s *IntroductionStrategy) Name() string {
	return "introduction"
}

func (s *IntroductionStrategy) Priority() int {
	return 80 // High priority for introductions
}

func (s *IntroductionStrategy) ShouldRespond(ctx context.Context, thread *models.Thread, messages []*models.Message, newMessage *models.Message) (bool, float64, error) {
	// Use LLM to analyze if the message is an introduction
	prompt := prompts.IntroductionAnalysisPrompt(newMessage)

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText("Analyze the message to determine if it's a user introduction. Be precise and return valid JSON.", genai.RoleModel),
		ResponseMIMEType:  "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"is_introduction": {
					Type: genai.TypeBoolean,
				},
				"confidence": {
					Type:    genai.TypeNumber,
					Minimum: &constants.MinimumConfidenceScore,
					Maximum: &constants.MaximumConfidenceScore,
				},
				"reasoning": {
					Type:      genai.TypeString,
					MaxLength: &constants.MaxAnalysisLength,
				},
			},
			Required: []string{"is_introduction", "confidence", "reasoning"},
		},
	}

	response, err := s.gemini.GenerateContent(ctx, prompt, config)

	s.logger.InfoContext(ctx, "Analyze introduction should respond strategy", "response", response)

	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to analyze introduction intent", "error", err)
		// Fallback to false on error
		return false, 0.0, err
	}

	// Parse the LLM response
	analysisResult, err := prompts.ParseIntroductionAnalysisResponse(response)
	if err != nil {
		s.logger.WarnContext(ctx, "Failed to parse introduction analysis response", "response", response, "error", err)
		// Fallback to false on parsing error
		return false, 0.0, nil
	}

	// Validate confidence bounds
	confidence := analysisResult.Confidence
	if confidence < 0.0 {
		confidence = 0.0
	}
	if confidence > 1.0 {
		confidence = 1.0
	}

	shouldRespond := analysisResult.IsIntroduction

	s.logger.InfoContext(ctx, "Introduction analysis completed",
		"is_introduction", analysisResult.IsIntroduction,
		"confidence", confidence,
		"reasoning", analysisResult.Reasoning,
		"should_respond", shouldRespond,
		"text_length", len(newMessage.Text))

	return shouldRespond, confidence, nil
}

func (s *IntroductionStrategy) GenerateResponse(ctx context.Context, thread *models.Thread, messages []*models.Message, newMessage *models.Message) (string, error) {
	// Extract user information from the message
	userInfo, err := s.extractUserInformation(ctx, newMessage)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to extract user information", "error", err)
		return "", err
	}

	// Update user in database
	err = s.updateUserProfile(ctx, newMessage, userInfo)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to update user profile", "error", err)
		// Don't return error - we can still generate a response
	}

	// Generate confirmation response
	response, err := s.generateConfirmationResponse(ctx, newMessage, userInfo)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to generate confirmation response", "error", err)
		return "", err
	}

	return response, nil
}

func (s *IntroductionStrategy) extractUserInformation(ctx context.Context, message *models.Message) (*models.UserInformation, error) {
	prompt := prompts.UserInformationExtractionPrompt(message)

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText("Extract user information in JSON format. Be precise and only extract explicitly mentioned information.", genai.RoleModel),
		ResponseMIMEType:  "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"bio": {
					Type:      genai.TypeString,
					MaxLength: &constants.MaxUserBioLength,
				},
				"interests": {
					Type: genai.TypeArray,
					Items: &genai.Schema{
						Type:      genai.TypeString,
						MaxLength: &constants.MaxUserInterestLength,
					},
				},
				"hobbies": {
					Type: genai.TypeArray,
					Items: &genai.Schema{
						Type:      genai.TypeString,
						MaxLength: &constants.MaxUserHobbyLength,
					},
				},
			},
		},
	}

	response, err := s.gemini.GenerateContent(ctx, prompt, config)
	if err != nil {
		return nil, err
	}

	// Parse the JSON response
	userInfo, err := prompts.ParseUserInformationResponse(response)
	if err != nil {
		s.logger.WarnContext(ctx, "Failed to parse user information response", "response", response, "error", err)
		// Return empty info rather than fail
		return &models.UserInformation{}, nil
	}

	s.logger.InfoContext(ctx, "Extracted user information",
		"bio_length", len(userInfo.Bio),
		"interests_count", len(userInfo.Interests),
		"hobbies_count", len(userInfo.Hobbies))

	return userInfo, nil
}

func (s *IntroductionStrategy) updateUserProfile(ctx context.Context, message *models.Message, userInfo *models.UserInformation) error {
	// Create or update basic user information
	err := s.userService.CreateOrUpdateUser(ctx, message.UserID, message.ChatID,
		message.FirstName, message.LastName, message.Username)
	if err != nil {
		return err
	}

	// Update bio if provided
	if userInfo.Bio != "" {
		err = s.userService.UpdateUserBio(ctx, message.UserID, userInfo.Bio)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to update user bio", "error", err)
		}
	}

	// Add interests if provided
	if len(userInfo.Interests) > 0 {
		err = s.userService.AddUserInterests(ctx, message.UserID, userInfo.Interests)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to add user interests", "error", err)
		}
	}

	// Add hobbies if provided
	if len(userInfo.Hobbies) > 0 {
		err = s.userService.AddUserHobbies(ctx, message.UserID, userInfo.Hobbies)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to add user hobbies", "error", err)
		}
	}

	return nil
}

func (s *IntroductionStrategy) generateConfirmationResponse(ctx context.Context, message *models.Message, userInfo *models.UserInformation) (string, error) {
	prompt := prompts.IntroductionConfirmationPrompt(message, userInfo)

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText("Generate a warm, friendly confirmation message. Keep it concise but personal. Maximum 300 characters.", genai.RoleModel),
		ResponseMIMEType:  "text/plain",
	}

	response, err := s.gemini.GenerateContent(ctx, prompt, config)
	if err != nil {
		return "", err
	}

	s.logger.InfoContext(ctx, "Generated confirmation response", "response_length", len(response))

	return response, nil
}
