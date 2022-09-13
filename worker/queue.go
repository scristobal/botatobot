package worker

import (
	"context"
	"scristobal/botatobot/cfg"
	"sync"
)

type Job interface {
	Run()
	Read() []byte
}

type CurrentJob struct {
	job *Job
	mut sync.RWMutex
}

var (
	pending chan Job
	done    chan Job
	current CurrentJob
)

func Init(ctx context.Context) {

	pending = make(chan Job, cfg.MAX_JOBS)

	done = make(chan Job, cfg.MAX_JOBS)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case job := <-pending:
				{
					process(job)
					done <- job
				}
			}
		}
	}()

}

func process(job Job) {

	current.mut.Lock()
	current.job = &job
	current.mut.Unlock()

	defer func() {
		current.mut.Lock()
		current.job = nil
		current.mut.Unlock()
	}()

	job.Run()
}

func Push(job Job) {
	pending <- job
}

func Pop() Job {
	return <-done
}

func Len() int {
	return len(pending)
}

func Current() *Job {
	current.mut.RLock()
	defer current.mut.RUnlock()
	return current.job
}
