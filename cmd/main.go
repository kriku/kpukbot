package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"regexp"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/kriku/kpukbot/app"
)

// Send any text message to the bot after the bot has been started
func main() {
	a, err := app.InitAppLocal()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	a.TelegramService.Bot.RegisterHandlerRegexp(
		bot.HandlerTypeMessageText,
		regexp.MustCompile(".*"),
		func(ctx context.Context, _ *bot.Bot, update *models.Update) {
			a.TelegramService.HandleUpdate(ctx, update)
		},
	)

	a.TelegramService.Bot.Start(ctx)
}

// func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
// 	b.SendMessage(ctx, &bot.SendMessageParams{
// 		ChatID: update.Message.Chat.ID,
// 		Text:   update.Message.Text,
// 	})
// }
