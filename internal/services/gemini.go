package services

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"github.com/kriku/kpukbot/internal/config"
	"google.golang.org/api/option"
)

// AIServiceInterface defines the interface for AI content generation
type AIServiceInterface interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
}

// GeminiService implements AIService using Google's Gemini API
type GeminiService struct {
	client *genai.Client
	model  string
}

// NewGeminiService creates a new GeminiService with the specified model
func NewGeminiService(c *config.Config) (AIServiceInterface, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(c.GeminiAPIKey))
	if err != nil {
		return &GeminiService{}, err
	}

	return &GeminiService{
		client: client,
		model:  c.GeminiModelName,
	}, nil
}

// GenerateContent implements AIService.GenerateContent
func (s *GeminiService) GenerateContent(ctx context.Context, prompt string) (string, error) {
	model := s.client.GenerativeModel(s.model)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	return readResponse(resp), nil
}

// Close properly closes the client
func (s *GeminiService) Close() {
	s.client.Close()
}

func readResponse(resp *genai.GenerateContentResponse) string {
	response := ""
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				response += fmt.Sprintln(part)
			}
		}
	}
	return response
}
