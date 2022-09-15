package controllers

import (
	"context"
	"fmt"
	"log"
	"scristobal/botatobot/config"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
)

func Generate[T request](ctx context.Context, b *bot.Bot, message models.Message, q Queue[T, models.Message]) {

	err := q.Push(message)

	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:           message.Chat.ID,
			Text:             fmt.Sprintf("Sorry, but your request is somehow invalid 😬\n\n %s", err),
			ReplyToMessageID: message.ID,
		})
		log.Printf("User %s requested %s but rejected", message.From.Username, err)
		return
	}

	id := uuid.New()

	log.Printf("User %s requested %s accepted\n", message.From.Username, id)

	if q.Len() >= config.MAX_JOBS {
		b.SendMessage(ctx,
			&bot.SendMessageParams{
				ChatID:           message.Chat.ID,
				Text:             "Sorry, but the job queue reached its maximum, try again later 🙄",
				ReplyToMessageID: message.ID,
			})

		log.Println("User", message.From.Username, "request rejected, queue full")
		return
	}

	log.Printf("User %s request accepted, job id %s", message.From.Username, id)

}
