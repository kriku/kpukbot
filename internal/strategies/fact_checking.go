package strategies

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/kriku/kpukbot/internal/clients/gemini"
	"github.com/kriku/kpukbot/internal/constants"
	"github.com/kriku/kpukbot/internal/models"
	"github.com/kriku/kpukbot/internal/prompts"
	"google.golang.org/genai"
)

type FactCheckingStrategy struct {
	gemini gemini.Client
	logger *slog.Logger
}

func NewFactCheckingStrategy(gemini gemini.Client, logger *slog.Logger) *FactCheckingStrategy {
	return &FactCheckingStrategy{
		gemini: gemini,
		logger: logger.With("strategy", "fact_checking"),
	}
}

func (s *FactCheckingStrategy) Name() string {
	return "fact_checking"
}

func (s *FactCheckingStrategy) Priority() int {
	return 80 // High priority
}

func (s *FactCheckingStrategy) ShouldRespond(ctx context.Context, thread *models.Thread, messages []*models.Message, newMessage *models.Message) (bool, float64, error) {
	// Check if message contains factual claims or questions
	text := strings.ToLower(newMessage.Text)

	// Keywords that might indicate fact-checking is needed
	keywords := []string{"is it true", "fact", "actually", "really", "correct", "wrong", "источник", "правда ли", "на самом деле"}

	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			s.logger.InfoContext(ctx, "Fact-checking keyword detected", "keyword", keyword)
			return true, 0.75, nil
		}
	}

	// Build context from recent messages
	var contextBuilder strings.Builder
	for _, msg := range messages {
		contextBuilder.WriteString(msg.Text)
		contextBuilder.WriteString(" ")
	}

	prompt := prompts.FactCheckingNeedsPrompt(contextBuilder.String(), newMessage.Text)

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"needs_checking": {Type: genai.TypeBoolean},
				"confidence": {
					Type:    genai.TypeNumber,
					Minimum: &constants.MinimumConfidenceScore,
					Maximum: &constants.MaximumConfidenceScore,
				},
			},
		},
	}

	response, err := s.gemini.GenerateContent(ctx, prompt, config)

	s.logger.InfoContext(ctx, "Fact checker analyzer response", "response", response)

	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to analyze for fact-checking", "error", err)
		return false, 0, err
	}

	var result struct {
		NeedsChecking bool    `json:"needs_checking"`
		Confidence    float64 `json:"confidence"`
	}

	// Try to parse JSON response
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		s.logger.WarnContext(ctx, "Failed to parse fact-checking analysis", "error", err)
		return false, 0, nil
	}

	return result.NeedsChecking, result.Confidence, nil
}

func (s *FactCheckingStrategy) GenerateResponse(ctx context.Context, thread *models.Thread, messages []*models.Message, newMessage *models.Message) (string, error) {
	// Build context from recent messages
	var contextBuilder strings.Builder
	for _, msg := range messages {
		contextBuilder.WriteString(msg.Text)
		contextBuilder.WriteString(" ")
	}

	prompt := prompts.FactCheckingPrompt(contextBuilder.String(), newMessage.Text)

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"verified":   {Type: genai.TypeBoolean},
				"confidence": {Type: genai.TypeNumber},
				"explanation": {
					Type:      genai.TypeString,
					MaxLength: &constants.MaxFactCheckingExplanationLength,
				},
				"additional_info": {
					Type:      genai.TypeString,
					MaxLength: &constants.MaxFactCheckingAdditionalInfoLength,
				},
			},
		},
	}

	response, err := s.gemini.GenerateContent(ctx, prompt, config)

	s.logger.InfoContext(ctx, "Fact checker response", "response", response)

	if err != nil {
		return "", err
	}

	// Parse the JSON response
	var result struct {
		Verified       bool    `json:"verified"`
		Confidence     float64 `json:"confidence"`
		Explanation    string  `json:"explanation"`
		AdditionalInfo string  `json:"additional_info"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// If parsing fails, return the raw response
		return "🔍 Факт:\n" + response, nil
	}

	// Format a nice response
	var responseBuilder strings.Builder
	responseBuilder.WriteString("🔍 Факт:\n\n")

	if result.Verified {
		responseBuilder.WriteString("✅ Абсолютно верно.\n\n")
	} else {
		responseBuilder.WriteString("⚠️ Возможно это не правда.\n\n")
	}

	responseBuilder.WriteString(result.Explanation)

	if result.AdditionalInfo != "" {
		responseBuilder.WriteString("\n\n📚 Additional context: ")
		responseBuilder.WriteString(result.AdditionalInfo)
	}

	return responseBuilder.String(), nil
}
