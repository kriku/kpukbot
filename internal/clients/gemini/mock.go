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
	case strings.Contains(promptLower, "analyze whether the following message is a user introduction"):
		return m.mockIntroductionAnalysisResponse(prompt)

	case strings.Contains(promptLower, "analyze the following message and determine"):
		return m.mockThreadClassificationResponse()

	case strings.Contains(promptLower, "create a brief summary for the following discussion thread"):
		return m.mockThreadSummaryResponse()

	case strings.Contains(promptLower, "you are an assistant bot analyzing whether a response is needed in this discussion"):
		return m.mockResponseAnalysisResponse()

	case strings.Contains(promptLower, "generate a helpful and contextually appropriate response"):
		return m.mockGeneralResponse()

	case strings.Contains(promptLower, "you are an ai assistant that determines whether a user's message should trigger an answer assessment response"):
		return m.mockShouldRespondEvaluationResponse(prompt)

	default:
		return m.mockGeneralResponse()
	}
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

// mockIntroductionAnalysisResponse generates appropriate mock responses for introduction analysis
func (m *MockClient) mockIntroductionAnalysisResponse(prompt string) string {
	promptLower := strings.ToLower(prompt)

	// Simple heuristic to determine if the message looks like an introduction
	// This mimics what the real LLM would do
	isIntroduction := false
	confidence := 0.2 // Default low confidence
	reasoning := "Does not contain typical introduction patterns"

	// Look for introduction indicators in the prompt (the message content)
	introPatterns := []string{
		"hello everyone! i'm", "hi, i'm", "my name is",
		"i am", "i'm", "about me", "introduce myself",
		"i like", "i love", "i enjoy", "i work", "passionate about",
		"nice to meet", "pleased to meet",
	}

	for _, pattern := range introPatterns {
		if strings.Contains(promptLower, pattern) {
			isIntroduction = true
			confidence = 0.8
			reasoning = "Contains self-referential language and personal information sharing typical of introductions"
			break
		}
	}

	// Simple greeting without personal info
	if strings.Contains(promptLower, "hi there!") || strings.Contains(promptLower, "hello!") {
		isIntroduction = false
		confidence = 0.1
		reasoning = "Simple greeting without personal information or self-disclosure"
	}

	// Questions or general conversation
	if strings.Contains(promptLower, "what") || strings.Contains(promptLower, "how") || strings.Contains(promptLower, "weather") {
		isIntroduction = false
		confidence = 0.05
		reasoning = "Appears to be a question or general conversation rather than an introduction"
	}

	return fmt.Sprintf(`{
		"is_introduction": %t,
		"confidence": %.2f,
		"reasoning": "%s"
	}`, isIntroduction, confidence, reasoning)
}

// mockShouldRespondEvaluationResponse generates mock responses for assessment strategy evaluation
func (m *MockClient) mockShouldRespondEvaluationResponse(prompt string) string {
	promptLower := strings.ToLower(prompt)

	// Analyze the conversation context to determine if assessment should respond
	shouldRespond := false
	confidence := 0.1
	reason := "No clear question-answer pattern detected"

	// Look for patterns that suggest this is an answer to a bot question
	if strings.Contains(promptLower, "bot:") && strings.Contains(promptLower, "user:") {
		// There's a bot-user conversation
		if strings.Contains(promptLower, "?") {
			// Bot asked a question
			shouldRespond = true
			confidence = 0.85
			reason = "User appears to be responding to a bot question"
		}
	}

	// Look for specific question patterns in the conversation
	questionPatterns := []string{
		"what's your favorite", "how do you", "why do you", "when did you",
		"where do you", "which", "who", "tell me about", "describe", "explain",
	}

	for _, pattern := range questionPatterns {
		if strings.Contains(promptLower, pattern) {
			shouldRespond = true
			confidence = 0.9
			reason = "Clear question-answer pattern detected in conversation"
			break
		}
	}

	// If user message is very short, lower confidence
	if strings.Count(promptLower, " ") < 5 {
		confidence = confidence * 0.5
		if confidence < 0.2 {
			shouldRespond = false
			reason = "User response too brief to warrant assessment"
		}
	}

	return fmt.Sprintf(`{
		"should_respond": %t,
		"confidence": %.2f,
		"reason": "%s"
	}`, shouldRespond, confidence, reason)
}
