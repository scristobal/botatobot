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

type task interface {
	Launch()
	Describe() string
	Result() ([]byte, error)
}

type Request[T task] struct {
	task *Runner[T]
	id   uuid.UUID
	msg  *models.Message
}

type Runner[T task] struct {
	Runner T
}

type Queue[T task] struct {
	requestFactory func(models.Message) ([]Request[T], error)
	current        *struct {
		req *Request[T]
		mut sync.RWMutex
	}
	pending chan Request[T]
	done    chan Request[T]
	bot     *bot.Bot
	ctx     context.Context
}

func NewQueue[T task](ctx context.Context, generator func(M models.Message) ([]*Runner[T], error)) Queue[T] {

	requestGenerator := func(m models.Message) ([]Request[T], error) {

		tasks, err := generator(m)

		if err != nil {

			return nil, fmt.Errorf("failed to create request: %s", err)

		}

		var requests []Request[T]

		for _, task := range tasks {
			task := task
			requests = append(requests, Request[T]{task, uuid.New(), &m})
		}

		return requests, nil
	}

	pending := make(chan Request[T], config.MAX_JOBS)

	done := make(chan Request[T], config.MAX_JOBS)

	current := &struct {
		req *Request[T]
		mut sync.RWMutex
	}{}

	return Queue[T]{requestGenerator, current, pending, done, nil, ctx}

}

func (q Queue[T]) Push(msg models.Message) error {
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

func (q Queue[T]) Pop() Request[T] {
	return <-q.done
}

func (q Queue[T]) Len() int {
	return len(q.pending)
}

func (q Queue[T]) IsWorking() bool {
	q.current.mut.RLock()
	defer q.current.mut.RUnlock()

	return q.current.req != nil
}

func (q *Queue[T]) RegisterBot(b *bot.Bot) {
	q.bot = b
}

func (queue Queue[T]) Start() {

	go func() {
		for {
			select {
			case <-queue.ctx.Done():
				return
			default:
				req := queue.Pop()

				_, err := req.task.Runner.Result()

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

					req.Job().Launch()
					queue.done <- req
				}()
			}
		}
	}()
	<-queue.ctx.Done()

}
