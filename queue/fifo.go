package queue

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"scristobal/botatobot/cfg"
	"strings"
	"sync"
	"time"
)

type serializable interface {
	ToJSON() ([]byte, error)
}

type Job struct {
	Id     string
	ChatId int
	User   string
	UserId int
	MsgId  int
	Params serializable
	Type   string
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

	rand.Seed(time.Now().UnixNano())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case job := <-pending:
				{
					job.process(ctx)
					done <- job
				}
			}
		}
	}()

}

func (job Job) process(ctx context.Context) {

	current.mut.Lock()
	current.job = &job
	current.mut.Unlock()

	defer func() {
		current.mut.Lock()
		current.job = nil
		current.mut.Unlock()
	}()

	type modelResponse struct {
		Status string   `json:"status"`
		Output []string `json:"output"` // (base64) data URLs
	}

	input, err := job.Params.ToJSON()

	if err != nil {
		log.Printf("error marshaling input: %v", err)
		return
	}

	res, err := http.Post(cfg.MODEL_URL, "application/json", strings.NewReader(fmt.Sprintf(`{"input": %s}`, input)))

	if err != nil {
		log.Printf("Error job %s while requesting model: %s\n", job.Id, err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		log.Printf("Error job %s while reading model response: %s\n", job.Id, err)
		return
	}

	response := modelResponse{}

	json.Unmarshal(body, &response)

	output := response.Output[0]

	// remove the data URL prefix
	data := strings.SplitAfter(output, ",")[1]

	decoded, err := base64.StdEncoding.DecodeString(data)

	if err != nil {
		log.Printf("Error job %s while decoding model response: %s\n", job.Id, err)
		return
	}

	imgFilePath := filepath.Join(cfg.OUTPUT_PATH, fmt.Sprintf("%s.png", job.Id))

	err = os.WriteFile(imgFilePath, decoded, 0644)

	if err != nil {
		log.Printf("Error job %s while writing image: %s\n", job.Id, err)
		return
	}

	content, err := json.Marshal(job)

	if err != nil {
		log.Printf("Error marshalling job %v", err)
		return
	}

	jsonFilePath := filepath.Join(cfg.OUTPUT_PATH, fmt.Sprintf("%s.json", job.Id))

	err = os.WriteFile(jsonFilePath, content, 0644)

	if err != nil {
		log.Printf("Error writing meta.json of job %s: %v", job.Id, err)
		return
	}

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
