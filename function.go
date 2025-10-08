package kpukbot

import (
	"log"
	"net/http"

	"github.com/kriku/kpukbot/app"
)

func HandleTelegramWebhook(res http.ResponseWriter, req *http.Request) {
	a, err := app.InitAppLocal()
	if err != nil {
		log.Printf("Failed to initialize application: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("initialization error"))
		return
	}
	defer a.Close()

	a.MessengerClient.HandleWebhook(res, req)

	res.WriteHeader(http.StatusOK)
	res.Write([]byte("ok"))
}
