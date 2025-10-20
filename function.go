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

			if json.Unmarshal(body, &triggerReq) == nil {
				switch triggerReq.Trigger {
				case "question":
					a.Logger.InfoContext(ctx, "Trigger question")
					handleQuestionTrigger(ctx, res, req, a)
					return
				case "rephrase-question":
					a.Logger.InfoContext(ctx, "Trigger rephrase question")
					handleRephraseQuestionTrigger(ctx, res, req, a)
					return
				}
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
			sentMessage, err := a.MessengerClient.SendMessage(ctx, chat.ID, question)
			if err != nil {
				log.Printf("Failed to send question to chat %d: %v", chat.ID, err)
			} else {
				questionsAsked++
				log.Printf("Asked question to user %d in chat %d", userID, chat.ID)

				// Save the question to the database and mark as asked with message ID
				if sentMessage != nil {
					err = questionStrategy.SaveQuestionAsMessage(ctx, chat.ID, sentMessage.ID, question)
					if err != nil {
						log.Printf("Failed to save question to database for chat %d: %v", chat.ID, err)
						// Don't fail the entire process if saving fails
					} else {
						log.Printf("Saved question to database with message ID %d", sentMessage.ID)
					}

					// Mark question as asked and save to user history
					err = questionStrategy.MarkQuestionAsAskedWithText(ctx, chat.ID, userID, sentMessage.ID, question)
					if err != nil {
						log.Printf("Failed to mark question as asked for user %d in chat %d: %v", userID, chat.ID, err)
						// Don't fail the entire process if marking fails
					}
				}
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

// handleRephraseQuestionTrigger handles the rephrase-question trigger request
func handleRephraseQuestionTrigger(ctx context.Context, res http.ResponseWriter, _ *http.Request, a app.App) {
	log.Printf("Processing rephrase question trigger request")

	// Find the question strategy
	questionStrategy := findQuestionStrategy(a)
	if questionStrategy == nil {
		log.Printf("Question strategy not found")
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("question strategy not found"))
		return
	}

	// Get all users currently in asking status
	usersInAsking, err := a.ChatsService.GetUsersInAskingStatus(ctx)
	if err != nil {
		log.Printf("Failed to get users in asking status: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("failed to get users in asking status"))
		return
	}

	questionsRephrased := 0
	for _, askingEntry := range usersInAsking {
		// Get user details through the question strategy
		user, err := questionStrategy.GetUser(ctx, askingEntry.UserID)
		if err != nil {
			log.Printf("Failed to get user %d: %v", askingEntry.UserID, err)
			continue
		}

		if user == nil {
			log.Printf("User %d not found", askingEntry.UserID)
			continue
		}

		// Get the original question using the stored question ID
		originalQuestion, err := questionStrategy.GetQuestionByID(ctx, askingEntry.QuestionID)
		if err != nil {
			log.Printf("Failed to get original question for user %d in chat %d with question ID %d: %v", askingEntry.UserID, askingEntry.ChatID, askingEntry.QuestionID, err)
			continue
		}

		// Rephrase the question
		rephrasedQuestion, err := questionStrategy.RephraseQuestionForUser(ctx, user, originalQuestion)
		if err != nil {
			log.Printf("Failed to rephrase question for user %d: %v", askingEntry.UserID, err)
			continue
		}

		// Send the rephrased question
		sentMessage, err := a.MessengerClient.SendMessage(ctx, askingEntry.ChatID, rephrasedQuestion)
		if err != nil {
			log.Printf("Failed to send rephrased question to user %d in chat %d: %v", askingEntry.UserID, askingEntry.ChatID, err)
		} else {
			questionsRephrased++
			log.Printf("Sent rephrased question to user %d in chat %d", askingEntry.UserID, askingEntry.ChatID)

			// Save the rephrased question to the database
			if sentMessage != nil {
				err = questionStrategy.SaveQuestionAsMessage(ctx, askingEntry.ChatID, sentMessage.ID, rephrasedQuestion)
				if err != nil {
					log.Printf("Failed to save rephrased question to database for chat %d: %v", askingEntry.ChatID, err)
					// Don't fail the entire process if saving fails
				} else {
					log.Printf("Saved rephrased question to database with message ID %d", sentMessage.ID)
				}

				// Update the question ID in the queue and save rephrased question to user history
				err = questionStrategy.MarkQuestionAsAskedWithText(ctx, askingEntry.ChatID, askingEntry.UserID, sentMessage.ID, rephrasedQuestion)
				if err != nil {
					log.Printf("Failed to update question ID for user %d in chat %d: %v", askingEntry.UserID, askingEntry.ChatID, err)
					// Don't fail the entire process if updating fails
				}
			}
		}
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(map[string]interface{}{
		"status":              "success",
		"questions_rephrased": questionsRephrased,
		"users_in_asking":     len(usersInAsking),
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
