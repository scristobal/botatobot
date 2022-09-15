package worker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"scristobal/botatobot/config"

	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
)

type Request struct {
	t   *Txt2img
	id  uuid.UUID       //`json:"id"`
	msg *models.Message //`json:"message"`
}

func New(m models.Message) ([]Request, error) {

	jobs, err := FromString(m.Text)

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}

	var requests []Request

	for _, job := range jobs {
		job := job
		requests = append(requests, Request{&job, uuid.New(), &m})
	}

	return requests, nil
}

func (r Request) Id() uuid.UUID {
	return r.id
}

func (r Request) Msg() *models.Message {
	return r.msg
}

func (r Request) Run() {
	r.t.Run()
}

func (r Request) Result() ([]byte, error) {
	return r.t.Result()
}

func (r Request) String() string {
	return r.t.String()
}

func (r *Request) MarshalJSON() ([]byte, error) {
	fmt.Println("marshalling request")
	return json.Marshal(struct {
		Id   uuid.UUID       `json:"id"`
		Msg  *models.Message `json:"message"`
		Task Txt2img         `json:"task"`
	}{
		r.id,
		r.msg,
		*r.t,
	})
}

func (r *Request) SaveToDisk() error {

	err := os.MkdirAll(config.OUTPUT_PATH, 0755)

	if err != nil {
		return fmt.Errorf("failed to create output directory: %s", err)
	}

	img, err := r.Result()

	if err != nil {
		return fmt.Errorf("failed to get result: %s", err)
	}

	imgFilePath := filepath.Join(config.OUTPUT_PATH, fmt.Sprintf("%s.png", r.Id()))

	err = os.WriteFile(imgFilePath, img, 0644)

	if err != nil {
		return fmt.Errorf("can't write image to disc: %s", err)
	}

	content, err := json.Marshal(r)

	if err != nil {
		return fmt.Errorf("failed to serialize job parameters: %s", err)
	}

	jsonFilePath := filepath.Join(config.OUTPUT_PATH, fmt.Sprintf("%s.json", r.Id()))

	err = os.WriteFile(jsonFilePath, content, 0644)

	if err != nil {
		return fmt.Errorf("error writing metadata: %s", err)
	}
	return nil
}
