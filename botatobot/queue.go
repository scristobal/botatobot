package botatobot

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"path/filepath"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
)

type Queue struct {
	factory func(models.Message) ([]Request, error)
	current *Request
	pending chan Request
	done    chan Request
	bot     *bot.Bot
	ctx     *context.Context
}

func NewQueue() Queue {
	requestGenerator := func(m models.Message) ([]Request, error) {

		tasks, err := Txt2imgFromString(m.Text)

		if err != nil {
			return nil, fmt.Errorf("failed to create request: %s", err)
		}

		var requests []Request

		u, err := url.Parse(MODEL_URL)

		if err != nil {
			return nil, fmt.Errorf("failed to parse model url: %s", err)
		}

		env := "remote"

		if u.Hostname() == "127.0.0.1" {
			env = "local"
		}

		for _, task := range tasks {
			task := task
			requests = append(requests, Request{*task, uuid.New(), &m, nil, nil, env})
		}

		return requests, nil
	}

	return Queue{
		factory: requestGenerator,
		pending: make(chan Request, MAX_JOBS),
		done:    make(chan Request, MAX_JOBS),
	}
}

func (q Queue) Push(msg models.Message) error {
	tasks, err := q.factory(msg)

	if err != nil {
		return fmt.Errorf("failed to create tasks: %s", err)
	}

	for _, task := range tasks {
		if len(q.pending) >= MAX_JOBS {
			return fmt.Errorf("too many jobs")
		}
		q.pending <- task
	}

	return nil
}

func (q Queue) Len() int {
	return len(q.pending)
}

func (q Queue) IsWorking() bool {
	return q.current != nil
}

func (q *Queue) RegisterBot(b *bot.Bot) {
	q.bot = b
}

func (q Queue) Start(ctx context.Context) {

	q.ctx = &ctx

	go func() {
		for {
			select {
			case <-(*q.ctx).Done():
				return
			case req := <-q.done:
				_, err := req.Result()

				if err != nil {
					log.Printf("Error processing request %s: %v", req.GetIdentifier(), err)
				}

				err = q.notifyBot(req)

				if err != nil {
					log.Printf("Error notifying user of %s: %v", req.GetIdentifier(), err)
				}

				err = req.SaveToDisk()
				if err != nil {
					log.Printf("Error saving request %s to disk: %v", req.GetIdentifier(), err)
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case <-(*q.ctx).Done():
				return
			case req := <-q.pending:
				q.current = &req

				req.Launch()
				q.done <- req

				q.current = nil
			}
		}
	}()
	<-(*q.ctx).Done()
}

func (q Queue) notifyBot(req Request) error {

	if q.bot == nil {
		return fmt.Errorf("bot not registered, use q.RegisterBot(b)")
	}

	if q.ctx == nil {
		return fmt.Errorf("context not registered, start the queue with q.Start(ctx)")
	}

	message := req.GetMessage()

	res, err := req.Result()

	if err != nil {

		_, err := q.bot.SendMessage(*q.ctx, &bot.SendMessageParams{
			ChatID:           message.Chat.ID,
			Text:             fmt.Sprintf("Sorry, but something went wrong 😭 %s", err),
			ReplyToMessageID: message.ID,
		})

		return err
	}

	_, err = q.bot.SendPhoto(*q.ctx, &bot.SendPhotoParams{
		ChatID:  message.Chat.ID,
		Caption: fmt.Sprint(&req.Task),
		Photo: &models.InputFileUpload{
			Data:     bytes.NewReader(res),
			Filename: filepath.Base(fmt.Sprintf("%s.png", req.GetIdentifier())),
		},
		DisableNotification: true,
	})

	return err
}
