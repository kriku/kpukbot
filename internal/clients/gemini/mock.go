package gemini

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"google.golang.org/genai"
)

// MockClient is a mock implementation of the Client interface for local testing
type MockClient struct {
	logger *slog.Logger
}

// NewMockClient creates a new mock Gemini client that doesn't make actual LLM requests
func NewMockClient(logger *slog.Logger) Client {
	return &MockClient{
		logger: logger.With("client", "gemini-mock"),
	}
}

func (m *MockClient) GenerateContent(ctx context.Context, prompt string, config *genai.GenerateContentConfig) (string, error) {
	m.logger.DebugContext(ctx, "Mock: Generating content", "prompt_length", len(prompt))

	// Generate a mock response based on the prompt content
	response := m.generateMockResponse(prompt)

	m.logger.DebugContext(ctx, "Mock: Content generated", "response_length", len(response))
	return response, nil
}

func (m *MockClient) GenerateContentWithHistory(ctx context.Context, history []Message, prompt string) (string, error) {
	m.logger.DebugContext(ctx, "Mock: Generating content with history",
		"history_length", len(history),
		"prompt_length", len(prompt))

	// Generate a mock response that acknowledges the history
	var historyContext strings.Builder
	if len(history) > 0 {
		historyContext.WriteString("Based on our conversation, ")
	}

	response := historyContext.String() + m.generateMockResponse(prompt)

	m.logger.DebugContext(ctx, "Mock: Content generated with history", "response_length", len(response))
	return response, nil
}

func (m *MockClient) Close() error {
	m.logger.Debug("Mock: Closing client")
	return nil
}

// generateMockResponse creates a contextual mock response based on the prompt
func (m *MockClient) generateMockResponse(prompt string) string {
	promptLower := strings.ToLower(prompt)

	// Check for specific patterns and return appropriate mock responses
	// Order matters - more specific patterns should come first
	switch {
	case strings.Contains(promptLower, "analyze the following message and determine"):
		return m.mockThreadClassificationResponse()

	case strings.Contains(promptLower, "create a brief summary for the following discussion thread"):
		return m.mockThreadSummaryResponse()

	case strings.Contains(promptLower, "you are an assistant bot analyzing whether a response is needed in this discussion"):
		return m.mockResponseAnalysisResponse()

	case strings.Contains(promptLower, "generate a helpful and contextually appropriate response"):
		return m.mockGeneralResponse()

	default:
		return m.mockGeneralResponse()
	}
}

// Mock responses for recent context analysis
func (m *MockClient) mockRecentContextAnalysisResponse() string {
	return `{
		"should_respond": true,
		"confidence": 0.75,
		"reasoning": "The conversation shows active engagement and the latest message appears to be seeking input or continuing a discussion.",
		"conversation_type": "discussion"
	}`
}

// Mock responses for recent context strategy selection
func (m *MockClient) mockRecentContextStrategySelectionResponse() string {
	return `{
		"approach": "conversational",
		"tone": "friendly",
		"focus": "addressing the user's question while maintaining the conversational flow"
	}`
}

// Mock responses for recent context response generation
func (m *MockClient) mockRecentContextResponse() string {
	return "This is a mock response generated with the recent context strategy. The response takes into account the conversational flow and aims to be helpful while maintaining an appropriate tone."
}

// Mock responses for thread classification
func (m *MockClient) mockThreadClassificationResponse() string {
	return `{
		"matches": [
			{
				"thread_id": "mock-thread-123",
				"probability": 0.75,
				"reasoning": "Message content is related to the ongoing discussion about testing strategies."
			}
		],
		"new_thread_suggestion": {
			"theme": "Mock Testing Discussion",
			"probability": 0.25
		}
	}`
}

// Mock responses for thread summary generation
func (m *MockClient) mockThreadSummaryResponse() string {
	return `{
		"theme": "Testing and Development",
		"summary": "Discussion about mock responses and testing strategies for the bot application."
	}`
}

// Mock responses for response analysis (ResponseAnalysisPrompt)
func (m *MockClient) mockResponseAnalysisResponse() string {
	return `{
		"should_respond": true,
		"confidence": 0.70,
		"reason": "The message appears to be directed at the bot and contains a question that would benefit from a response.",
		"suggested_strategy": "general"
	}`
}

func (m *MockClient) mockGeneralResponse() string {
	return fmt.Sprintf(`This is a mock response from the Gemini client simulator.

Your request has been processed locally without making an actual LLM API call. This is useful for:
- Development and testing
- Reducing API costs
- Working offline
- Fast iteration cycles

The actual response would contain more contextual and relevant information based on your specific prompt.

Mock timestamp: %s`, "2025-10-13T00:00:00Z")
}
