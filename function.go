package kpukbot

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/kriku/kpukbot/app"
	"github.com/kriku/kpukbot/internal/strategies"
)

// TriggerRequest represents the structure for custom trigger requests
type TriggerRequest struct {
	Trigger string `json:"trigger"`
}

func HandleTelegramWebhook(res http.ResponseWriter, req *http.Request) {
	ctx := context.Background()

	a, err := app.InitApp(ctx)
	if err != nil {
		log.Printf("Failed to initialize application: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("initialization error"))
		return
	}
	defer a.Close()


	// Check if this is a custom trigger request
	if req.Method == "POST" && req.Header.Get("Content-Type") == "application/json" {
		body, err := io.ReadAll(req.Body)

		a.Logger.InfoContext(ctx, "Handling incoming request", "body", body)

		if err == nil {
			var triggerReq TriggerRequest

			if json.Unmarshal(body, &triggerReq) == nil && triggerReq.Trigger == "question" {
				a.Logger.InfoContext(ctx, "Trigger question")
				handleQuestionTrigger(ctx, res, req, a)
				return
			}
		}
		// Reset body for telegram webhook handling
		req.Body = io.NopCloser(strings.NewReader(string(body)))
	}

	// Handle as telegram webhook
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("ok"))

	handleCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	a.MessengerClient.HandleWebhook(handleCtx, res, req)
}

// handleQuestionTrigger handles the question trigger request
func handleQuestionTrigger(ctx context.Context, res http.ResponseWriter, req *http.Request, a app.App) {
	log.Printf("Processing question trigger request")

	// Get all chats
	chats, err := a.ChatsService.GetAllChats(ctx)
	if err != nil {
		log.Printf("Failed to get all chats: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("failed to get chats"))
		return
	}

	// Process each chat
	questionsAsked := 0
	for _, chat := range chats {
		if !chat.IsActive {
			continue
		}

		// Find the question strategy (we'll need to access it from the orchestrator)
		questionStrategy := findQuestionStrategy(a)
		if questionStrategy == nil {
			log.Printf("Question strategy not found")
			continue
		}

		// Ask question to the next user in queue for this chat
		question, userID, err := questionStrategy.AskQuestionToUser(ctx, chat.ID)
		if err != nil {
			log.Printf("Failed to ask question in chat %d: %v", chat.ID, err)
			continue
		}

		if userID > 0 && question != "" {
			// Send the question to the chat using the messenger client
			_, err = a.MessengerClient.SendMessage(ctx, chat.ID, question)
			if err != nil {
				log.Printf("Failed to send question to chat %d: %v", chat.ID, err)
			} else {
				questionsAsked++
				log.Printf("Asked question to user %d in chat %d", userID, chat.ID)
			}
		}
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(map[string]interface{}{
		"status":          "success",
		"questions_asked": questionsAsked,
		"chats_processed": len(chats),
	})
}

// findQuestionStrategy finds the question strategy from the available strategies
func findQuestionStrategy(a app.App) *strategies.QuestionStrategy {
	for _, strategy := range a.Strategies {
		if strategy.Name() == "question" {
			if questionStrategy, ok := strategy.(*strategies.QuestionStrategy); ok {
				return questionStrategy
			}
		}
	}
	return nil
}
