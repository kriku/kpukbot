package kpukbot

import (
	"context"
	"log"
	"net/http"

	"github.com/kriku/kpukbot/app"
)

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

	a.MessengerClient.HandleWebhook(ctx, res, req)

	res.WriteHeader(http.StatusOK)
	res.Write([]byte("ok"))
}
