package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	pkg "scristobal/botatobot/pkg"

	telegrambot "github.com/go-telegram/bot"
)

func main() {

	if err := pkg.FromEnv(); err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	queue := pkg.NewQueue()

	bot, err := telegrambot.New(pkg.BOT_TOKEN)

	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	bot.RegisterHandler(telegrambot.HandlerTypeMessageText, string(pkg.GenerateCmd), telegrambot.MatchTypePrefix, pkg.Generate(queue))
	bot.RegisterHandler(telegrambot.HandlerTypeMessageText, string(pkg.StatusCmd), telegrambot.MatchTypePrefix, pkg.Status(queue))
	bot.RegisterHandler(telegrambot.HandlerTypeMessageText, string(pkg.HelpCmd), telegrambot.MatchTypePrefix, pkg.Help())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	callback := pkg.SendOutcome(ctx, bot)

	queue.SetCallback(callback)

	go queue.Start(ctx)
	go bot.Start(ctx)

	log.Println("Bot online, listening to messages...")

	go pkg.Start_health()

	<-ctx.Done()
}
