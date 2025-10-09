package strategies

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/kriku/kpukbot/internal/clients/gemini"
	"github.com/kriku/kpukbot/internal/models"
	"github.com/kriku/kpukbot/internal/prompts"
	"google.golang.org/genai"
)

type AgreementStrategy struct {
	gemini gemini.Client
	logger *slog.Logger
}

func NewAgreementStrategy(gemini gemini.Client, logger *slog.Logger) *AgreementStrategy {
	return &AgreementStrategy{
		gemini: gemini,
		logger: logger.With("strategy", "agreement"),
	}
}

func (s *AgreementStrategy) Name() string {
	return "agreement"
}

func (s *AgreementStrategy) Priority() int {
	return 75 // High priority
}

func (s *AgreementStrategy) ShouldRespond(ctx context.Context, thread *models.Thread, messages []*models.Message, newMessage *models.Message) (bool, float64, error) {
	text := strings.ToLower(newMessage.Text)

	// Keywords indicating agreements or decisions
	keywords := []string{"agree", "decided", "let's do", "consensus", "согласны", "решили", "договорились", "deal", "ok", "okay"}

	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			s.logger.InfoContext(ctx, "Agreement keyword detected", "keyword", keyword)
			return true, 0.7, nil
		}
	}

	return false, 0, nil
}

func (s *AgreementStrategy) GenerateResponse(ctx context.Context, thread *models.Thread, messages []*models.Message, newMessage *models.Message) (string, error) {
	prompt := prompts.AgreementTrackingPrompt(thread, append(messages, newMessage))

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"topic":    {Type: genai.TypeString},
					"decision": {Type: genai.TypeString},
					"participants": {
						Type: genai.TypeArray,
						Items: &genai.Schema{
							Type: genai.TypeString,
						},
					},
					"confidence": {Type: genai.TypeNumber},
				},
			},
		},
	}

	response, err := s.gemini.GenerateContent(ctx, prompt, config)

	s.logger.InfoContext(ctx, "Agreement generate response", "response", response)

	if err != nil {
		return "", err
	}

	var result struct {
		Agreements []struct {
			Topic        string   `json:"topic"`
			Decision     string   `json:"decision"`
			Participants []string `json:"participants"`
			Confidence   float64  `json:"confidence"`
		} `json:"agreements"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return "📝 Agreement noted:\n" + response, nil
	}

	if len(result.Agreements) == 0 {
		return "", nil
	}

	var responseBuilder strings.Builder
	responseBuilder.WriteString("📝 Agreement(s) recorded:\n\n")

	for i, agreement := range result.Agreements {
		responseBuilder.WriteString("✓ **")
		responseBuilder.WriteString(agreement.Topic)
		responseBuilder.WriteString("**\n")
		responseBuilder.WriteString("   Decision: ")
		responseBuilder.WriteString(agreement.Decision)

		if len(agreement.Participants) > 0 {
			responseBuilder.WriteString("\n   Participants: ")
			responseBuilder.WriteString(strings.Join(agreement.Participants, ", "))
		}

		if i < len(result.Agreements)-1 {
			responseBuilder.WriteString("\n\n")
		}
	}

	return responseBuilder.String(), nil
}
