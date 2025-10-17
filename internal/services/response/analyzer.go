package response

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/kriku/kpukbot/internal/clients/gemini"
	"github.com/kriku/kpukbot/internal/constants"
	"github.com/kriku/kpukbot/internal/models"
	"github.com/kriku/kpukbot/internal/prompts"
	"github.com/kriku/kpukbot/internal/strategies"
	"google.golang.org/genai"
)

type AnalyzerService struct {
	gemini     gemini.Client
	strategies []strategies.ResponseStrategy
	logger     *slog.Logger
}

func NewAnalyzerService(
	gemini gemini.Client,
	strategies []strategies.ResponseStrategy,
	logger *slog.Logger,
) *AnalyzerService {
	return &AnalyzerService{
		gemini:     gemini,
		strategies: strategies,
		logger:     logger.With("service", "response_analyzer"),
	}
}

func (s *AnalyzerService) AnalyzeAndRespond(
	ctx context.Context,
	thread *models.Thread,
	messages []*models.Message,
	newMessage *models.Message,
) (string, error) {
	s.logger.InfoContext(ctx, "Analyzing response need",
		"thread_id", thread.ID,
		"message_count", len(messages))

	// First, use LLM to get general assessment
	prompt := prompts.ResponseAnalysisPrompt(thread, messages, newMessage)
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"should_respond": {Type: genai.TypeBoolean},
				"confidence": {
					Type:    genai.TypeNumber,
					Minimum: &constants.MinimumConfidenceScore,
					Maximum: &constants.MaximumConfidenceScore,
				},
				"reason": {
					Type:      genai.TypeString,
					MaxLength: &constants.MaxAnalysisLength,
				},
				"suggested_strategy": {
					Type: genai.TypeString,
					Enum: []string{"general", "introduction"},
				},
			},
		},
	}

	response, err := s.gemini.GenerateContent(ctx, prompt, config)

	s.logger.InfoContext(ctx, "Analyze and respond response", "response", response)

	if err != nil {
		s.logger.WarnContext(ctx, "Failed to get LLM analysis", "error", err)
	}

	var analysis struct {
		ShouldRespond     bool    `json:"should_respond"`
		Confidence        float64 `json:"confidence"`
		Reason            string  `json:"reason"`
		SuggestedStrategy string  `json:"suggested_strategy"`
	}

	if err := json.Unmarshal([]byte(response), &analysis); err != nil {
		s.logger.WarnContext(ctx, "Failed to parse LLM analysis", "error", err)
		analysis.ShouldRespond = false
	}

	if !analysis.ShouldRespond {
		s.logger.InfoContext(ctx, "LLM suggests no response needed", "reason", analysis.Reason)
		return "", nil
	}

	s.logger.InfoContext(ctx, "LLM suggests response",
		"confidence", analysis.Confidence,
		"suggested_strategy", analysis.SuggestedStrategy)

	// First, try to find and prioritize the suggested strategy
	var suggestedStrategy strategies.ResponseStrategy
	for _, strategy := range s.strategies {
		if strategy.Name() == analysis.SuggestedStrategy {
			suggestedStrategy = strategy
			break
		}
	}

	var bestResult *strategies.StrategyResult

	// If we found the suggested strategy, evaluate it first with bonus confidence
	if suggestedStrategy != nil {
		shouldRespond, confidence, err := suggestedStrategy.ShouldRespond(ctx, thread, messages, newMessage)
		if err != nil {
			s.logger.ErrorContext(ctx, "Suggested strategy evaluation failed",
				"strategy", suggestedStrategy.Name(),
				"error", err)
		} else if shouldRespond {
			// Give suggested strategy a significant confidence boost
			adjustedConfidence := confidence * (float64(suggestedStrategy.Priority()) / 100.0)
			bonusConfidence := adjustedConfidence + 0.3 // Add 30% bonus for being LLM-suggested
			if bonusConfidence > 1.0 {
				bonusConfidence = 1.0
			}

			s.logger.InfoContext(ctx, "Suggested strategy evaluated",
				"strategy", suggestedStrategy.Name(),
				"original_confidence", confidence,
				"adjusted_confidence", adjustedConfidence,
				"bonus_confidence", bonusConfidence)

			bestResult = &strategies.StrategyResult{
				Strategy:      suggestedStrategy,
				ShouldRespond: true,
				Confidence:    bonusConfidence,
			}
		}
	}

	// Evaluate remaining strategies only if suggested strategy didn't qualify or doesn't exist
	for _, strategy := range s.strategies {
		// Skip if this is the suggested strategy we already evaluated
		if suggestedStrategy != nil && strategy.Name() == analysis.SuggestedStrategy {
			continue
		}

		shouldRespond, confidence, err := strategy.ShouldRespond(ctx, thread, messages, newMessage)
		if err != nil {
			s.logger.ErrorContext(ctx, "Strategy evaluation failed",
				"strategy", strategy.Name(),
				"error", err)
			continue
		}

		if !shouldRespond {
			continue
		}

		// Adjust confidence based on priority
		adjustedConfidence := confidence * (float64(strategy.Priority()) / 100.0)

		s.logger.DebugContext(ctx, "Strategy suggests response",
			"strategy", strategy.Name(),
			"confidence", confidence,
			"adjusted_confidence", adjustedConfidence)

		if bestResult == nil || adjustedConfidence > bestResult.Confidence {
			bestResult = &strategies.StrategyResult{
				Strategy:      strategy,
				ShouldRespond: true,
				Confidence:    adjustedConfidence,
			}
		}
	}

	if bestResult == nil {
		s.logger.InfoContext(ctx, "No strategy suggests response")
		return "", nil
	}

	s.logger.InfoContext(ctx, "Selected strategy",
		"strategy", bestResult.Strategy.Name(),
		"confidence", bestResult.Confidence)

	// Generate response using the selected strategy
	responseText, err := bestResult.Strategy.GenerateResponse(ctx, thread, messages, newMessage)
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	return responseText, nil
}
