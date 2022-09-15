package queue

import (
	"context"
	"fmt"
	"scristobal/botatobot/config"
	"sync"
)

type job interface {
	Run()
}

type Queue[T job, M any] struct {
	gen     func(M) ([]T, error)
	current *struct {
		job *T
		mut sync.RWMutex
	}
	pending chan T
	done    chan T
}

type Generator[T job, M any] func(M) ([]T, error)

func New[T job, M any](ctx context.Context, generator Generator[T, M]) Queue[T, M] {

	pending := make(chan T, config.MAX_JOBS)

	done := make(chan T, config.MAX_JOBS)

	current := &struct {
		job *T
		mut sync.RWMutex
	}{}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case job := <-pending:
				{
					current.mut.Lock()
					current.job = &job
					current.mut.Unlock()

					defer func() {
						current.mut.Lock()
						current.job = nil
						current.mut.Unlock()
					}()

					job.Run()
					done <- job
				}
			}
		}
	}()

	return Queue[T, M]{generator, current, pending, done}

}

func (q Queue[T, M]) Push(item M) error {
	jobs, err := q.gen(item)

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

func (q Queue[T, M]) Pop() T {
	return <-q.done
}

func (q Queue[T, M]) Len() int {
	return len(q.pending)
}

func (q Queue[T, M]) IsWorking() bool {
	q.current.mut.RLock()
	defer q.current.mut.RUnlock()
	return q.current.job != nil
}
