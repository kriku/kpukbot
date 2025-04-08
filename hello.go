package hello

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

func HelloWorld(res http.ResponseWriter, req *http.Request) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}

	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}

	update, err := bot.HandleUpdate(req)
	if err != nil {
		log.Printf("Error handling update: %s", err)
		return
	}

	response := generateContent(update.Message.Text)

	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, response))

	res.WriteHeader(http.StatusOK)
	res.Write([]byte("ok"))
}
