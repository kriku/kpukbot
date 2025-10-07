package kpukbot

import (
	"log"
	"net/http"

	"github.com/kriku/kpukbot/app"
)

func HandleTelegramWebhook(res http.ResponseWriter, req *http.Request) {
	a, err := app.InitAppLocal()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	a.MessengerClient.HandleWebhook(res, req)

	res.WriteHeader(http.StatusOK)
	res.Write([]byte("ok"))
}
