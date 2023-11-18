package botatobot

import (
	"context"
	"fmt"
	"log"

	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
)

type Queue struct {
	factory  func(models.Message) ([]Request, error)
	current  *Request
	pending  chan Request
	done     chan Request
	callback func(Outcome) error
	ctx      *context.Context
}

func NewQueue() Queue {
	requestGenerator := func(m models.Message) ([]Request, error) {

		tasks, err := TaskFromString(m.Text)

		if err != nil {
			return nil, fmt.Errorf("failed to create request: %s", err)
		}

		var requests []Request

		for _, task := range tasks {
			task := task
			requests = append(requests, Request{Task: *task, Id: uuid.New(), Msg: &m})
		}

		return requests, nil
	}

	return Queue{
		factory: requestGenerator,
		pending: make(chan Request, MAX_JOBS),
		done:    make(chan Request, MAX_JOBS),
	}
}

func (q *Queue) Push(msg *models.Message) error {
	tasks, err := q.factory(*msg)

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

func (q *Queue) Len() int {
	return len(q.pending)
}

func (q *Queue) IsWorking() bool {
	return q.current != nil
}

func (q *Queue) Start(ctx context.Context) {

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

				if q.callback != nil {
					err = q.callback(&req)

					if err != nil {
						log.Printf("Error running callback of  %s: %v", req.GetIdentifier(), err)
					}
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

func (q *Queue) SetCallback(f func(Outcome) error) {
	q.callback = f
}
