package kpukbot

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/kriku/kpukbot/app"
)

func HandleTelegramWebhook(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Minute)
	defer cancel()

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
