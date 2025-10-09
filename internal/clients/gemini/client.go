package gemini

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/genai"
)

type Client interface {
	GenerateContent(ctx context.Context, prompt string, config *genai.GenerateContentConfig) (string, error)
	GenerateContentWithHistory(ctx context.Context, history []Message, prompt string) (string, error)
	Close() error
}

type Message struct {
	Role    string // "user" or "model"
	Content string
}

type GeminiClient struct {
	client *genai.Client
	logger *slog.Logger
}

func NewGeminiClient(ctx context.Context, apiKey string, logger *slog.Logger) (Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiClient{
		client: client,
		logger: logger.With("client", "gemini"),
	}, nil
}

func (g *GeminiClient) GenerateContent(ctx context.Context, prompt string, config *genai.GenerateContentConfig) (string, error) {
	g.logger.DebugContext(ctx, "Generating content", "prompt_length", len(prompt))

	// Create content from the prompt
	contents := genai.Text(prompt)

	// Use default config if none provided
	if config == nil {
		config = &genai.GenerateContentConfig{
			Temperature:     genai.Ptr(float32(0.7)),
			TopK:            genai.Ptr(float32(40)),
			TopP:            genai.Ptr(float32(0.95)),
			MaxOutputTokens: 2048,
		}
	}

	resp, err := g.client.Models.GenerateContent(ctx, "gemini-2.0-flash", contents, config)
	if err != nil {
		g.logger.ErrorContext(ctx, "Failed to generate content", "error", err)
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	result := resp.Text()
	g.logger.DebugContext(ctx, "Content generated", "response_length", len(result))

	return result, nil
}

func (g *GeminiClient) GenerateContentWithHistory(ctx context.Context, history []Message, prompt string) (string, error) {
	g.logger.DebugContext(ctx, "Generating content with history",
		"history_length", len(history),
		"prompt_length", len(prompt))

	// Convert history to genai.Content format
	var contents []*genai.Content
	for _, msg := range history {
		role := "user"
		if msg.Role == "model" {
			role = "model"
		}
		contents = append(contents, &genai.Content{
			Parts: []*genai.Part{genai.NewPartFromText(msg.Content)},
			Role:  role,
		})
	}

	// Create a chat session with history
	chat, err := g.client.Chats.Create(ctx, "gemini-2.5-flash", &genai.GenerateContentConfig{
		Temperature:     genai.Ptr(float32(0.7)),
		TopK:            genai.Ptr(float32(40)),
		TopP:            genai.Ptr(float32(0.95)),
		MaxOutputTokens: 2048,
	}, contents)
	if err != nil {
		g.logger.ErrorContext(ctx, "Failed to create chat session", "error", err)
		return "", fmt.Errorf("failed to create chat session: %w", err)
	}

	// Send the new message
	resp, err := chat.SendMessage(ctx, *genai.NewPartFromText(prompt))
	if err != nil {
		g.logger.ErrorContext(ctx, "Failed to send message in chat", "error", err)
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	result := resp.Text()
	g.logger.DebugContext(ctx, "Content generated with history", "response_length", len(result))

	return result, nil
}

func (g *GeminiClient) Close() error {
	// The new genai.Client doesn't require explicit closing
	return nil
}
