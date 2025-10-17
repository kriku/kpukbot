package kpukbot

import (
	"context"
	"log"
	"net/http"
	"time"

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

	res.WriteHeader(http.StatusOK)
	res.Write([]byte("ok"))

	handleCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	a.MessengerClient.HandleWebhook(handleCtx, res, req)
}
