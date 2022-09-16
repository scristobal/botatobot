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

type job interface {
	Run()
	Describe() string
	Result() ([]byte, error)
}

type request[T job] interface {
	Job() T
	Id() uuid.UUID
	Msg() *models.Message
	Result() ([]byte, error)
	SaveToDisk() error
}

type Queue[R request[T], T job] struct {
	requestFactory func(models.Message) ([]R, error)
	current        *struct {
		req *R
		mut sync.RWMutex
	}
	pending chan R
	done    chan R
	bot     *bot.Bot
	ctx     context.Context
}

type generator[T job] func(M models.Message) ([]T, error)

func NewQueue[T job](ctx context.Context, generator generator[T]) Queue[Request[T], T] {

	requestGenerator := func(m models.Message) ([]Request[T], error) {

		jobs, err := generator(m)

		if err != nil {

			return nil, fmt.Errorf("failed to create request: %s", err)

		}

		var requests []Request[T]

		for _, job := range jobs {
			job := job
			requests = append(requests, Request[T]{&job, uuid.New(), &m})
		}

		return requests, nil
	}

	pending := make(chan Request[T], config.MAX_JOBS)

	done := make(chan Request[T], config.MAX_JOBS)

	current := &struct {
		req *Request[T]
		mut sync.RWMutex
	}{}

	return Queue[Request[T], T]{requestGenerator, current, pending, done, nil, ctx}

}

func (q Queue[R, T]) Push(msg models.Message) error {
	jobs, err := q.requestFactory(msg)

	if err != nil {
		return fmt.Errorf("failed to create job: %s", err)
	}

	if len(jobs) > config.MAX_JOBS {
		return fmt.Errorf("too many jobs")
	}

	for _, job := range jobs {
		q.pending <- job
	}

	return nil
}

func (q Queue[R, T]) Pop() R {
	return <-q.done
}

func (q Queue[R, T]) Len() int {
	return len(q.pending)
}

func (q Queue[R, T]) IsWorking() bool {
	q.current.mut.RLock()
	defer q.current.mut.RUnlock()

	return q.current.req != nil
}

func (q *Queue[R, T]) RegisterBot(b *bot.Bot) {
	q.bot = b
}

func (queue Queue[R, T]) Start() func() {
	return func() {
		go func() {
			for {
				select {
				case <-queue.ctx.Done():
					return
				default:
					req := queue.Pop()

					_, err := req.Result()

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
					{
						queue.current.mut.Lock()
						queue.current.req = &req
						queue.current.mut.Unlock()

						defer func() {
							queue.current.mut.Lock()
							queue.current.req = nil
							queue.current.mut.Unlock()
						}()

						req.Job().Run()
						queue.done <- req
					}
				}
			}
		}()
	}

}
