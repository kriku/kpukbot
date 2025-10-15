package gemini

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestMockClient_GenerateContent(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	client := NewMockClient(logger)

	tests := []struct {
		name           string
		prompt         string
		expectedSubstr string
	}{
		{
			name:           "General query",
			prompt:         "Tell me about the weather",
			expectedSubstr: "mock response",
		},
		{
			name:           "Fact checking needs query",
			prompt:         "Does this fact needs checking?",
			expectedSubstr: `"needs_checking"`,
		},
		{
			name:           "Fact checking verification query",
			prompt:         "Please verify this fact about history",
			expectedSubstr: `"verified"`,
		},
		{
			name:           "Thread classification query",
			prompt:         "Classify this message for thread classification",
			expectedSubstr: `"matches"`,
		},
		{
			name:           "Thread summary query",
			prompt:         "Generate thread summary for this conversation",
			expectedSubstr: `"theme"`,
		},
		{
			name:           "Should respond analysis query",
			prompt:         "Analyze if should_respond to this message",
			expectedSubstr: `"should_respond"`,
		},
		{
			name:           "Strategy selection query",
			prompt:         "Select strategy selection approach",
			expectedSubstr: `"approach"`,
		},
		{
			name:           "Greeting",
			prompt:         "Hello, how are you?",
			expectedSubstr: "Hello!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			response, err := client.GenerateContent(ctx, tt.prompt, nil)

			if err != nil {
				t.Errorf("GenerateContent() error = %v", err)
				return
			}

			if response == "" {
				t.Error("GenerateContent() returned empty response")
				return
			}

			if !strings.Contains(response, tt.expectedSubstr) {
				t.Errorf("GenerateContent() response does not contain expected substring.\nExpected substring: %s\nGot: %s", tt.expectedSubstr, response)
			}

			t.Logf("Response: %s", response)
		})
	}
}

func TestMockClient_GenerateContentWithHistory(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	client := NewMockClient(logger)

	ctx := context.Background()
	history := []Message{
		{Role: "user", Content: "What is the capital of France?"},
		{Role: "model", Content: "The capital of France is Paris."},
		{Role: "user", Content: "What about Germany?"},
	}

	response, err := client.GenerateContentWithHistory(ctx, history, "Tell me more about it")

	if err != nil {
		t.Errorf("GenerateContentWithHistory() error = %v", err)
		return
	}

	if response == "" {
		t.Error("GenerateContentWithHistory() returned empty response")
		return
	}

	if !strings.Contains(response, "Based on our conversation") {
		t.Error("GenerateContentWithHistory() response does not acknowledge conversation history")
	}

	t.Logf("Response: %s", response)
}

func TestMockClient_Close(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	client := NewMockClient(logger)

	err := client.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestMockClient_JSONSchemaValidation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	client := NewMockClient(logger).(*MockClient)

	t.Run("Fact check needs response", func(t *testing.T) {
		response := client.mockFactCheckNeedsResponse()

		var result struct {
			NeedsChecking bool    `json:"needs_checking"`
			Confidence    float64 `json:"confidence"`
		}

		if err := json.Unmarshal([]byte(response), &result); err != nil {
			t.Errorf("Failed to parse fact check needs response: %v", err)
		}

		if result.Confidence < 0.0 || result.Confidence > 1.0 {
			t.Errorf("Confidence should be between 0.0 and 1.0, got %f", result.Confidence)
		}
	})

	t.Run("Thread classification response", func(t *testing.T) {
		response := client.mockThreadClassificationResponse()

		var result struct {
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

		if err := json.Unmarshal([]byte(response), &result); err != nil {
			t.Errorf("Failed to parse thread classification response: %v", err)
		}

		if len(result.Matches) == 0 {
			t.Error("Expected at least one match in thread classification response")
		}

		for _, match := range result.Matches {
			if match.ThreadID == "" {
				t.Error("Thread ID should not be empty")
			}
			if match.Probability < 0.0 || match.Probability > 1.0 {
				t.Errorf("Match probability should be between 0.0 and 1.0, got %f", match.Probability)
			}
		}
	})

	t.Run("Should respond analysis", func(t *testing.T) {
		response := client.mockShouldRespondResponse()

		var result struct {
			ShouldRespond    bool    `json:"should_respond"`
			Confidence       float64 `json:"confidence"`
			Reasoning        string  `json:"reasoning"`
			ConversationType string  `json:"conversation_type"`
			DetectedIntent   string  `json:"detected_intent"`
		}

		if err := json.Unmarshal([]byte(response), &result); err != nil {
			t.Errorf("Failed to parse should respond response: %v", err)
		}

		if result.Confidence < 0.0 || result.Confidence > 1.0 {
			t.Errorf("Confidence should be between 0.0 and 1.0, got %f", result.Confidence)
		}

		if result.Reasoning == "" {
			t.Error("Reasoning should not be empty")
		}
	})
}
