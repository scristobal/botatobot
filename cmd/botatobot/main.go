package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"scristobal/botatobot/pkg"

	"github.com/go-telegram/bot"
)

func main() {

	if err := pkg.FromEnv(); err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	botato, err := bot.New(pkg.TELEGRAMBOT_TOKEN)

	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	botato.RegisterHandler(bot.HandlerTypeMessageText, string(pkg.GenerateCommand), bot.MatchTypePrefix, pkg.GenerateHandler)
	botato.RegisterHandler(bot.HandlerTypeMessageText, string(pkg.HelpCommand), bot.MatchTypePrefix, pkg.HelpHandler)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	generator := pkg.NewImageGenerator()
	ctx = context.WithValue(ctx, pkg.ImageGeneratorKey, generator)

	go botato.Start(ctx)

	log.Println("Bot online, listening to messages...")

	if pkg.LOCAL_PORT != "" {
		log.Printf("Starting health check server on port %s\n", pkg.LOCAL_PORT)
		go pkg.StartHealthCheckServer()
	}

	<-ctx.Done()
}
