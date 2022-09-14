package worker

import (
	"context"
	"scristobal/botatobot/cfg"
	"sync"
)

type job interface {
	Run()
}

type Queue[T job] struct {
	current *struct {
		job *T
		mut sync.RWMutex
	}
	pending chan T
	done    chan T
}

func Init[T job](ctx context.Context) Queue[T] {

	pending := make(chan T, cfg.MAX_JOBS)

	done := make(chan T, cfg.MAX_JOBS)

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

	return Queue[T]{current, pending, done}

}

func (q *Queue[T]) Push(job T) {
	q.pending <- job
}

func (q *Queue[T]) Pop() T {
	return <-q.done
}

func (q *Queue[T]) Len() int {
	return len(q.pending)
}

func (q *Queue[T]) Current() *T {
	q.current.mut.RLock()
	defer q.current.mut.RUnlock()
	return q.current.job
}
