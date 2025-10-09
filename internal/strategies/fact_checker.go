package strategies

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/kriku/kpukbot/internal/clients/gemini"
	"github.com/kriku/kpukbot/internal/models"
	"github.com/kriku/kpukbot/internal/prompts"
)

type FactCheckerStrategy struct {
	gemini gemini.Client
	logger *slog.Logger
}

func NewFactCheckerStrategy(gemini gemini.Client, logger *slog.Logger) *FactCheckerStrategy {
	return &FactCheckerStrategy{
		gemini: gemini,
		logger: logger.With("strategy", "fact_checker"),
	}
}

func (s *FactCheckerStrategy) Name() string {
	return "fact_checking"
}

func (s *FactCheckerStrategy) Priority() int {
	return 80 // High priority
}

func (s *FactCheckerStrategy) ShouldRespond(ctx context.Context, thread *models.Thread, messages []*models.Message, newMessage *models.Message) (bool, float64, error) {
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

	// Use LLM to determine if fact-checking is needed
	prompt := "Does this message contain factual claims that should be verified? Answer with JSON: {\"needs_checking\": true/false, \"confidence\": 0.0-1.0}\n\nMessage: " + newMessage.Text

	config := &genai.GenerationConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"needs_checking": {Type: genai.TypeBoolean},
				"confidence":     {Type: genai.TypeNumber},
			},
		},
	}

	response, err := s.gemini.GenerateContent(ctx, prompt, config)
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

func (s *FactCheckerStrategy) GenerateResponse(ctx context.Context, thread *models.Thread, messages []*models.Message, newMessage *models.Message) (string, error) {
	// Build context from recent messages
	var contextBuilder strings.Builder
	for _, msg := range messages {
		contextBuilder.WriteString(msg.Text)
		contextBuilder.WriteString(" ")
	}

	prompt := prompts.FactCheckingPrompt(contextBuilder.String(), newMessage.Text)

	config := &genai.GenerationConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"verified":        {Type: genai.TypeBoolean},
				"confidence":      {Type: genai.TypeNumber},
				"explanation":     {Type: genai.TypeString},
				"additional_info": {Type: genai.TypeString},
			},
		},
	}

	response, err := s.gemini.GenerateContent(ctx, prompt, config)
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
		return "🔍 Fact Check:\n" + response, nil
	}

	// Format a nice response
	var responseBuilder strings.Builder
	responseBuilder.WriteString("🔍 Fact Check:\n\n")

	if result.Verified {
		responseBuilder.WriteString("✅ This appears to be correct.\n\n")
	} else {
		responseBuilder.WriteString("⚠️ This may not be accurate.\n\n")
	}

	responseBuilder.WriteString(result.Explanation)

	if result.AdditionalInfo != "" {
		responseBuilder.WriteString("\n\n📚 Additional context: ")
		responseBuilder.WriteString(result.AdditionalInfo)
	}

	return responseBuilder.String(), nil
}
