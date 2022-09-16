package controllers

import (
	"context"
	"fmt"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type msgPusher interface {
	Push(item models.Message) error
}

func Generate(ctx context.Context, b *bot.Bot, message models.Message, q msgPusher) {

	err := q.Push(message)

	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:           message.Chat.ID,
			Text:             fmt.Sprintf("Sorry, but your request was rejected 😬\n\n %s", err),
			ReplyToMessageID: message.ID,
		})
		log.Printf("User %s requested %s but rejected", message.From.Username, err)
		return
	}

}
