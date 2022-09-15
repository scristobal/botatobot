package controllers

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type request interface {
	Run()
	//Result() ([]byte, error)
}

type Queue[T request, M any] interface {
	Push(item M) error
	Len() int
	Current() *T
}

func Status[T request](ctx context.Context, b *bot.Bot, message models.Message, q Queue[T, models.Message]) {
	job := q.Current()

	numJobs := q.Len()

	if job == nil && numJobs == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: message.Chat.ID,
			Text:   "I am doing nothing and there are no jobs in the queue 🤖",
		})
		return
	}

	if job != nil {

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: message.Chat.ID,
			Text:   fmt.Sprintf("I am generating an image and the queue has %d more jobs", numJobs),
		})
		return
	}

	if numJobs > 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: message.Chat.ID,
			Text:   fmt.Sprintf("I am doing nothing and the queue has %d more jobs. That's weird!! ", numJobs),
		})
		return
	}

}
