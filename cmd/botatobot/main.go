package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"scristobal/botatobot/config"
	botatobot "scristobal/botatobot/pkg"
	"scristobal/botatobot/pkg/commands"
	"scristobal/botatobot/pkg/handlers"
	"scristobal/botatobot/pkg/server"
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

	b.RegisterHandler(bot.HandlerTypeMessageText, string(commands.Generate), bot.MatchTypePrefix, handlers.Generate(&queue))
	b.RegisterHandler(bot.HandlerTypeMessageText, string(commands.Status), bot.MatchTypePrefix, handlers.Status(&queue))
	b.RegisterHandler(bot.HandlerTypeMessageText, string(commands.Help), bot.MatchTypePrefix, handlers.Help())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	queue.SetCallback(botatobot.SendOutcome(ctx, b))

	go queue.Start(ctx)
	go b.Start(ctx)

	log.Println("Bot online, listening to messages...")

	go server.Http(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	<-ctx.Done()
}
