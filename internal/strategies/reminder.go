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

type ReminderStrategy struct {
	gemini gemini.Client
	logger *slog.Logger
}

func NewReminderStrategy(gemini gemini.Client, logger *slog.Logger) *ReminderStrategy {
	return &ReminderStrategy{
		gemini: gemini,
		logger: logger.With("strategy", "reminder"),
	}
}

func (s *ReminderStrategy) Name() string {
	return "reminder"
}

func (s *ReminderStrategy) Priority() int {
	return 70 // Medium-high priority
}

func (s *ReminderStrategy) ShouldRespond(ctx context.Context, thread *models.Thread, messages []*models.Message, newMessage *models.Message) (bool, float64, error) {
	text := strings.ToLower(newMessage.Text)

	// Keywords that might indicate reminders are needed
	keywords := []string{"remind", "deadline", "tomorrow", "next week", "don't forget", "remember", "напомни", "завтра", "не забудь", "срок"}

	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			s.logger.InfoContext(ctx, "Reminder keyword detected", "keyword", keyword)
			return true, 0.8, nil
		}
	}

	return false, 0, nil
}

func (s *ReminderStrategy) GenerateResponse(ctx context.Context, thread *models.Thread, messages []*models.Message, newMessage *models.Message) (string, error) {
	prompt := prompts.ReminderPrompt(thread, append(messages, newMessage))

	config := &genai.GenerationConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"reminders": {
					Type: genai.TypeArray,
					Items: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"person":   {Type: genai.TypeString},
							"action":   {Type: genai.TypeString},
							"deadline": {Type: genai.TypeString},
							"priority": {Type: genai.TypeString},
						},
					},
				},
			},
		},
	}

	response, err := s.gemini.GenerateContent(ctx, prompt, config)
	if err != nil {
		return "", err
	}

	var result struct {
		Reminders []struct {
			Person   string `json:"person"`
			Action   string `json:"action"`
			Deadline string `json:"deadline"`
			Priority string `json:"priority"`
		} `json:"reminders"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return "⏰ Reminder noted:\n" + response, nil
	}

	if len(result.Reminders) == 0 {
		return "", nil // No reminders found
	}

	var responseBuilder strings.Builder
	responseBuilder.WriteString("⏰ Reminders tracked:\n\n")

	for i, reminder := range result.Reminders {
		icon := "📌"
		if reminder.Priority == "high" {
			icon = "🔥"
		}

		responseBuilder.WriteString(icon)
		responseBuilder.WriteString(" ")
		if reminder.Person != "" {
			responseBuilder.WriteString(reminder.Person)
			responseBuilder.WriteString(": ")
		}
		responseBuilder.WriteString(reminder.Action)
		if reminder.Deadline != "" {
			responseBuilder.WriteString(" (by ")
			responseBuilder.WriteString(reminder.Deadline)
			responseBuilder.WriteString(")")
		}

		if i < len(result.Reminders)-1 {
			responseBuilder.WriteString("\n")
		}
	}

	return responseBuilder.String(), nil
}
