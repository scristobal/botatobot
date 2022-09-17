package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"scristobal/botatobot/botatobot"
	"time"

	"github.com/go-telegram/bot"
)

func main() {

	rand.Seed(time.Now().UnixNano())

	if err := botatobot.FromEnv(); err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	queue := botatobot.NewQueue()
	handler := botatobot.NewHandler(queue)

	opts := []bot.Option{bot.WithDefaultHandler(handler)}
	b := bot.New(botatobot.BOT_TOKEN, opts...)

	queue.RegisterBot(b)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go queue.Start(ctx)
	go b.Start(ctx)

	log.Println("Bot online, listening to messages...")

	<-ctx.Done()
}
