package strategies

import (
	"context"
	"log/slog"

	"github.com/kriku/kpukbot/internal/clients/gemini"
	"github.com/kriku/kpukbot/internal/models"
	"github.com/kriku/kpukbot/internal/prompts"
	"google.golang.org/genai"
)

type GeneralStrategy struct {
	gemini gemini.Client
	logger *slog.Logger
}

func NewGeneralStrategy(gemini gemini.Client, logger *slog.Logger) *GeneralStrategy {
	return &GeneralStrategy{
		gemini: gemini,
		logger: logger.With("strategy", "general"),
	}
}

func (s *GeneralStrategy) Name() string {
	return "general"
}

func (s *GeneralStrategy) Priority() int {
	return 30 // Lowest priority - fallback
}

func (s *GeneralStrategy) ShouldRespond(ctx context.Context, thread *models.Thread, messages []*models.Message, newMessage *models.Message) (bool, float64, error) {
	// General strategy is always available as fallback
	// But with low confidence to let other strategies take precedence
	return true, 0.3, nil
}

func (s *GeneralStrategy) GenerateResponse(ctx context.Context, thread *models.Thread, messages []*models.Message, newMessage *models.Message) (string, error) {
	prompt := prompts.GeneralResponsePrompt(thread, messages, newMessage)

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText("The maximum length of the answer is 4096 characters.", genai.RoleModel),
		ResponseMIMEType:  "text/plain",
	}

	response, err := s.gemini.GenerateContent(ctx, prompt, config)

	s.logger.InfoContext(ctx, "General response", "response", response)

	if err != nil {
		return "", err
	}

	return response, nil
}
