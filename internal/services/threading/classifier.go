package threading

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/kriku/kpukbot/internal/clients/gemini"
	"github.com/kriku/kpukbot/internal/constants"
	"github.com/kriku/kpukbot/internal/models"
	"github.com/kriku/kpukbot/internal/prompts"
	messagesRepo "github.com/kriku/kpukbot/internal/repository/messages"
	threadsRepo "github.com/kriku/kpukbot/internal/repository/threads"
	"google.golang.org/genai"
)

type ClassifierService struct {
	gemini                gemini.Client
	threadsRepo           threadsRepo.ThreadsRepository
	messagesRepo          messagesRepo.MessagesRepository
	logger                *slog.Logger
	minProbability        float64 // Minimum probability to consider a match
	sameUserTimeThreshold time.Duration
}

func NewClassifierService(
	gemini gemini.Client,
	threadsRepo threadsRepo.ThreadsRepository,
	messagesRepo messagesRepo.MessagesRepository,
	logger *slog.Logger,
) *ClassifierService {
	return &ClassifierService{
		gemini:                gemini,
		threadsRepo:           threadsRepo,
		messagesRepo:          messagesRepo,
		logger:                logger.With("service", "classifier"),
		minProbability:        0.5, // Default threshold
		sameUserTimeThreshold: 5 * time.Minute,
	}
}

func (s *ClassifierService) ClassifyMessage(ctx context.Context, message *models.Message) (*models.ThreadMatch, error) {
	s.logger.InfoContext(ctx, "Classifying message", "message_id", message.ID, "chat_id", message.ChatID)

	// Handle replies
	if message.ReplyToMessageID != 0 {
		thread, err := s.threadsRepo.GetThreadByMessageID(ctx, message.ReplyToMessageID)
		if err != nil {
			s.logger.WarnContext(ctx, "Failed to get thread by reply message ID", "error", err)
		}
		if thread != nil {
			s.logger.InfoContext(ctx, "Message is a reply, adding to existing thread", "thread_id", thread.ID)
			err := s.AddMessageToThread(ctx, thread, message)
			if err != nil {
				return nil, fmt.Errorf("failed to add message to thread: %w", err)
			}
			return &models.ThreadMatch{
				Thread:      thread,
				Probability: 1.0,
				Reasoning:   "Direct reply to a message in the thread",
			}, nil
		}
	}

	// Get active threads for this chat
	threads, err := s.threadsRepo.GetActiveThreadsByChatID(ctx, message.ChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get threads: %w", err)
	}

	// If no active threads, create a new one
	if len(threads) == 0 {
		s.logger.InfoContext(ctx, "No active threads found, creating new thread")
		return s.createNewThread(ctx, message)
	}

	// Use LLM to classify the message
	prompt := prompts.ThreadClassificationPrompt(message, threads)

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"matches": {
					Type: genai.TypeArray,
					Items: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"thread_id":   {Type: genai.TypeString},
							"probability": {Type: genai.TypeNumber},
							"reasoning": {
								Type:      genai.TypeString,
								MaxLength: &constants.MaxThreadReasoningLength,
							},
						},
					},
				},
				"new_thread_suggestion": {
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"theme": {
							Type:      genai.TypeString,
							MaxLength: &constants.MaxThreadThemeLength,
						},
						"probability": {Type: genai.TypeNumber},
					},
				},
			},
		},
	}

	response, err := s.gemini.GenerateContent(ctx, prompt, config)

	s.logger.InfoContext(ctx, "Analyzer classification response", "response", response)

	if err != nil {
		return nil, fmt.Errorf("failed to classify message: %w", err)
	}

	// Parse the classification result
	var classification struct {
		Matches []struct {
			ThreadID    string  `json:"thread_id"`
			Probability float64 `json:"probability"`
			Reasoning   string  `json:"reasoning"`
		} `json:"matches"`
		NewThreadSuggestion struct {
			Theme       string  `json:"theme"`
			Probability float64 `json:"probability"`
		} `json:"new_thread_suggestion"`
	}

	if err := json.Unmarshal([]byte(response), &classification); err != nil {
		s.logger.WarnContext(ctx, "Failed to parse classification response", "error", err)
		// Fallback: create new thread
		return s.createNewThread(ctx, message)
	}

	// Find the best match
	var bestMatch *models.ThreadMatch
	for _, match := range classification.Matches {
		if match.Probability >= s.minProbability {
			// Find the thread
			for _, thread := range threads {
				if thread.ID == match.ThreadID {
					if bestMatch == nil || match.Probability > bestMatch.Probability {
						bestMatch = &models.ThreadMatch{
							Thread:      thread,
							Probability: match.Probability,
							Reasoning:   match.Reasoning,
						}
					}
					break
				}
			}
		}
	}

	// If we found a good match, use it
	if bestMatch != nil {
		s.logger.InfoContext(ctx, "Found matching thread",
			"thread_id", bestMatch.Thread.ID,
			"probability", bestMatch.Probability)
		return bestMatch, nil
	}

	// Otherwise, create a new thread
	s.logger.InfoContext(ctx, "No matching thread found, creating new one")
	return s.createNewThread(ctx, message)
}

func (s *ClassifierService) createNewThread(ctx context.Context, message *models.Message) (*models.ThreadMatch, error) {
	// Generate theme and summary for the new thread
	messages := []*models.Message{message}
	prompt := prompts.ThreadSummaryPrompt(messages)

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"theme": {
					Type:      genai.TypeString,
					MaxLength: &constants.MaxThreadThemeLength,
				},
				"summary": {
					Type:      genai.TypeString,
					MaxLength: &constants.MaxThreadSummaryLength,
				},
			},
		},
	}

	response, err := s.gemini.GenerateContent(ctx, prompt, config)

	s.logger.InfoContext(ctx, "Analyzer create new thread response", "response", response)

	if err != nil {
		s.logger.WarnContext(ctx, "Failed to generate thread summary", "error", err)
		// Use fallback
		response = fmt.Sprintf(`{"theme": "New conversation", "summary": "%s"}`, message.Text[:min(50, len(message.Text))])
	}

	var summary struct {
		Theme   string `json:"theme"`
		Summary string `json:"summary"`
	}

	if err := json.Unmarshal([]byte(response), &summary); err != nil {
		summary.Theme = "New conversation"
		summary.Summary = message.Text[:min(100, len(message.Text))]
	}

	thread := &models.Thread{
		ID:         uuid.New().String(),
		ChatID:     message.ChatID,
		Theme:      summary.Theme,
		Summary:    summary.Summary,
		MessageIDs: []int{message.ID},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		IsActive:   true,
	}

	if err := s.threadsRepo.SaveThread(ctx, thread); err != nil {
		return nil, fmt.Errorf("failed to save new thread: %w", err)
	}

	s.logger.InfoContext(ctx, "Created new thread", "thread_id", thread.ID, "theme", thread.Theme)

	return &models.ThreadMatch{
		Thread:      thread,
		Probability: 1.0,
		Reasoning:   "New thread created",
	}, nil
}

func (s *ClassifierService) AddMessageToThread(ctx context.Context, thread *models.Thread, message *models.Message) error {
	// Add message ID to thread
	thread.MessageIDs = append(thread.MessageIDs, message.ID)
	thread.UpdatedAt = time.Now()

	// Update thread summary periodically (every 5 messages)
	if len(thread.MessageIDs)%5 == 0 {
		if err := s.updateThreadSummary(ctx, thread); err != nil {
			s.logger.WarnContext(ctx, "Failed to update thread summary", "error", err)
		}
	}

	return s.threadsRepo.UpdateThread(ctx, thread)
}

func (s *ClassifierService) updateThreadSummary(ctx context.Context, thread *models.Thread) error {
	// Get recent messages from the thread
	var messages []*models.Message
	// Get last 10 messages
	startIdx := max(0, len(thread.MessageIDs)-10)
	for i := startIdx; i < len(thread.MessageIDs); i++ {
		msgList, err := s.messagesRepo.GetMessage(ctx, int64(thread.MessageIDs[i]))
		if err != nil {
			continue
		}
		messages = append(messages, msgList...)
	}

	if len(messages) == 0 {
		return nil
	}

	prompt := prompts.ThreadSummaryPrompt(messages)

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"theme": {
					Type:      genai.TypeString,
					MaxLength: &constants.MaxThreadThemeLength,
				},
				"summary": {
					Type:      genai.TypeString,
					MaxLength: &constants.MaxThreadSummaryLength,
				},
			},
		},
	}

	response, err := s.gemini.GenerateContent(ctx, prompt, config)

	s.logger.InfoContext(ctx, "Analyzer update thread summary response", "response", response)

	if err != nil {
		return err
	}

	var summary struct {
		Theme   string `json:"theme"`
		Summary string `json:"summary"`
	}

	if err := json.Unmarshal([]byte(response), &summary); err != nil {
		return err
	}

	thread.Theme = summary.Theme
	thread.Summary = summary.Summary

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
