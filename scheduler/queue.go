package scheduler

import (
	"context"
	"fmt"
	"log"
	"scristobal/botatobot/config"
	"scristobal/botatobot/handlers"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
)

type Request struct {
	task Txt2img
	id   uuid.UUID
	msg  *models.Message
}

type Queue struct {
	requestFactory func(models.Message) ([]Request, error)
	current        *struct {
		req *Request
		mut sync.RWMutex
	}
	pending chan Request
	done    chan Request
	bot     *bot.Bot
	ctx     context.Context
}

func NewQueue(ctx context.Context) Queue {

	requestGenerator := func(m models.Message) ([]Request, error) {

		requestFactory := func(m models.Message) ([]*Txt2img, error) {
			return FromString(m.Text)

		}

		tasks, err := requestFactory(m)

		if err != nil {

			return nil, fmt.Errorf("failed to create request: %s", err)

		}

		var requests []Request

		for _, task := range tasks {
			task := task
			requests = append(requests, Request{*task, uuid.New(), &m})
		}

		return requests, nil
	}

	pending := make(chan Request, config.MAX_JOBS)

	done := make(chan Request, config.MAX_JOBS)

	current := &struct {
		req *Request
		mut sync.RWMutex
	}{}

	return Queue{requestGenerator, current, pending, done, nil, ctx}

}

func (q Queue) Push(msg models.Message) error {
	tasks, err := q.requestFactory(msg)

	if err != nil {
		return fmt.Errorf("failed to create tasks: %s", err)
	}

	if len(tasks) > config.MAX_JOBS {
		return fmt.Errorf("too many jobs")
	}

	for _, task := range tasks {
		q.pending <- task
	}

	return nil
}

func (q Queue) Pop() Request {
	return <-q.done
}

func (q Queue) Len() int {
	return len(q.pending)
}

func (q Queue) IsWorking() bool {
	q.current.mut.RLock()
	defer q.current.mut.RUnlock()

	return q.current.req != nil
}

func (q *Queue) RegisterBot(b *bot.Bot) {
	q.bot = b
}

func (queue Queue) Start() {

	go func() {
		for {
			select {
			case <-queue.ctx.Done():
				return
			default:
				req := queue.Pop()

				_, err := req.task.Result()

				if err != nil {
					log.Printf("Error processing request %s: %v", req.Id(), err)
				}

				err = handlers.Request(queue.ctx, queue.bot, req)

				if err != nil {
					log.Printf("Error notifying user of %s: %v", req.Id(), err)
				}

				err = req.SaveToDisk()
				if err != nil {
					log.Printf("Error saving request %s to disk: %v", req.Id(), err)
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case <-queue.ctx.Done():
				return
			case req := <-queue.pending:
				func() {
					queue.current.mut.Lock()
					queue.current.req = &req
					queue.current.mut.Unlock()

					defer func() {
						queue.current.mut.Lock()
						queue.current.req = nil
						queue.current.mut.Unlock()
					}()

					req.task.Launch()
					queue.done <- req
				}()
			}
		}
	}()
	<-queue.ctx.Done()

}
