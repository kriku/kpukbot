package hello

import (
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text))

	res.WriteHeader(http.StatusOK)
	res.Write([]byte("ok"))
}
