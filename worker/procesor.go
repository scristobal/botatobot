package worker

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
	"scristobal/botatobot/cfg"
	"strings"
	"sync"
	"time"
)

var (
	pending chan Job
	done    chan Job
	current currentJob
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

	type modelResponse struct {
		Status string   `json:"status"`
		Output []string `json:"output"` // (base64) data URLs
	}

	outputFolder := fmt.Sprintf("%s/%s", cfg.OUTPUT_PATH, job.Id)

	err := os.MkdirAll(outputFolder, 0755)

	if err != nil {
		return
	}

	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {

		wg.Add(1)

		go func() {

			defer wg.Done()

			seed := rand.Intn(1000000)

			res, err := http.Post(cfg.MODEL_URL, "application/json", strings.NewReader(fmt.Sprintf(`{"input": {"prompt": "%s","seed": %d}}`, job.Prompt, seed)))

			if err != nil {
				log.Printf("Error job %s while requesting model: %s\n", job.Id, err)
			}

			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)

			if err != nil {
				log.Printf("Error job %s while reading model response: %s\n", job.Id, err)
			}

			response := modelResponse{}

			json.Unmarshal(body, &response)

			output := response.Output[0]

			// remove the data URL prefix
			data := strings.SplitAfter(output, ",")[1]

			decoded, err := base64.StdEncoding.DecodeString(data)

			if err != nil {
				log.Printf("Error job %s while decoding model response: %s\n", job.Id, err)
			}

			fileName := fmt.Sprintf("seed_%d.png", seed)

			filePath := fmt.Sprintf("%s/%s", outputFolder, fileName)

			err = os.WriteFile(filePath, decoded, 0644)

			if err != nil {
				log.Printf("Error job %s while writing image: %s\n", job.Id, err)
			}
		}()
	}

	wg.Wait()

	content, err := json.Marshal(job)

	if err != nil {
		log.Printf("Error marshalling job %v", err)
	}

	err = os.WriteFile(fmt.Sprintf("%s/meta.json", outputFolder), content, 0644)

	if err != nil {
		log.Printf("Error writing meta.json of job %s: %v", job.Id, err)
	}

	current.mut.Lock()
	current.job = nil
	current.mut.Unlock()

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
