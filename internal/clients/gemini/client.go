package gemini

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Client interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
	GenerateContentWithHistory(ctx context.Context, history []Message, prompt string) (string, error)
	Close() error
}

type Message struct {
	Role    string // "user" or "model"
	Content string
}

type GeminiClient struct {
	client *genai.Client
	model  *genai.GenerativeModel
	logger *slog.Logger
}

func NewGeminiClient(ctx context.Context, apiKey string, logger *slog.Logger) (Client, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	model := client.GenerativeModel("gemini-2.5-flash")

	// Configure model settings
	model.SetTemperature(0.7)
	model.SetTopK(40)
	model.SetTopP(0.95)
	model.SetMaxOutputTokens(2048)

	return &GeminiClient{
		client: client,
		model:  model,
		logger: logger.With("client", "gemini"),
	}, nil
}

func (g *GeminiClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	g.logger.DebugContext(ctx, "Generating content", "prompt_length", len(prompt))

	resp, err := g.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		g.logger.ErrorContext(ctx, "Failed to generate content", "error", err)
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	result := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	g.logger.DebugContext(ctx, "Content generated", "response_length", len(result))

	return result, nil
}

func (g *GeminiClient) GenerateContentWithHistory(ctx context.Context, history []Message, prompt string) (string, error) {
	g.logger.DebugContext(ctx, "Generating content with history",
		"history_length", len(history),
		"prompt_length", len(prompt))

	// Start a chat session
	cs := g.model.StartChat()

	// Add history
	for _, msg := range history {
		role := msg.Role
		if role != "user" && role != "model" {
			role = "user" // Default to user if invalid
		}

		cs.History = append(cs.History, &genai.Content{
			Parts: []genai.Part{genai.Text(msg.Content)},
			Role:  role,
		})
	}

	// Send the new message
	resp, err := cs.SendMessage(ctx, genai.Text(prompt))
	if err != nil {
		g.logger.ErrorContext(ctx, "Failed to send message in chat", "error", err)
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	result := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	g.logger.DebugContext(ctx, "Content generated with history", "response_length", len(result))

	return result, nil
}

func (g *GeminiClient) Close() error {
	return g.client.Close()
}
