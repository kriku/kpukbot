package hello

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

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

func generateContent(prompt string) string {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Fatal(err)
	}

	return readResponse(resp)
}

func HandleTelegramWebhook(res http.ResponseWriter, req *http.Request) {
	b, err := bot.New(os.Getenv("TELEGRAM_API_TOKEN"))
	if err != nil {
		panic(err)
	}

	update := models.Update{}

	err = json.NewDecoder(req.Body).Decode(&update)
	if err != nil {
		log.Printf("Error parsing incoming webhook update: %s", err)
	}

	HandleUpdate(b, &update)

	res.WriteHeader(http.StatusOK)
	res.Write([]byte("ok"))
}

func HandleUpdate(b *bot.Bot, update *models.Update) {
	ctx := context.Background()

	log.Printf("Handle new message: [%s] %s", update.Message.From.Username, update.Message.Text)

	b.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: update.Message.Chat.ID,
		Action: models.ChatActionTyping,
	})

	response := generateContent(update.Message.Text)

	log.Printf("Generated content: %s", response)

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      bot.EscapeMarkdown(response),
		ParseMode: models.ParseModeMarkdown,
	})

	if err != nil {
		log.Printf("Error sending message: %s", err)
	}
}
