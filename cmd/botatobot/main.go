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

	bot, err := telegrambot.New(pkg.TELEGRAMBOT_TOKEN)

	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	bot.RegisterHandler(telegrambot.HandlerTypeMessageText, string(pkg.GenerateCmd), telegrambot.MatchTypePrefix, pkg.GenerateHandler)
	bot.RegisterHandler(telegrambot.HandlerTypeMessageText, string(pkg.HelpCmd), telegrambot.MatchTypePrefix, pkg.HelpHandler)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go bot.Start(ctx)

	log.Println("Bot online, listening to messages...")

	if pkg.LOCAL_PORT != "" {
		log.Printf("Starting health check server on port %s\n", pkg.LOCAL_PORT)
		go pkg.Start_health()
	}

	<-ctx.Done()
}
