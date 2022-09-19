package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"scristobal/botatobot/config"
	botatobot "scristobal/botatobot/pkg"
	"time"

	"github.com/go-telegram/bot"
)

func main() {

	rand.Seed(time.Now().UnixNano())

	if err := config.FromEnv(); err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	queue := botatobot.NewQueue()

	b := bot.New(config.BOT_TOKEN)

	b.RegisterHandler(bot.HandlerTypeMessageText, string(botatobot.Generate), bot.MatchTypePrefix, botatobot.GenerateHandler(&queue))
	b.RegisterHandler(bot.HandlerTypeMessageText, string(botatobot.Status), bot.MatchTypePrefix, botatobot.StatusHandler(&queue))
	b.RegisterHandler(bot.HandlerTypeMessageText, string(botatobot.Help), bot.MatchTypePrefix, botatobot.HelpHandler())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	queue.SetCallback(botatobot.SendRequest(ctx, b))

	go queue.Start(ctx)
	go b.Start(ctx)

	log.Println("Bot online, listening to messages...")

	<-ctx.Done()
}
