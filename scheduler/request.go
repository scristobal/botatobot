package scheduler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"scristobal/botatobot/config"

	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
)

type Request[T job] struct {
	task *T
	id   uuid.UUID
	msg  *models.Message
}

func (r Request[T]) Id() uuid.UUID {
	return r.id
}

func (r Request[T]) Msg() *models.Message {
	return r.msg
}

func (r Request[T]) Result() ([]byte, error) {
	return (*r.task).Result()

}

func (r Request[T]) Run() {
	(*r.task).Run()
}

func (r Request[T]) Describe() string {
	return (*r.task).Describe()
}

func (r Request[T]) Job() T {
	return *r.task
}

func (r Request[T]) String() string {
	return (*r.task).Describe()
}

func (r Request[T]) MarshalJSON() ([]byte, error) {
	fmt.Println("marshalling request")
	return json.Marshal(struct {
		Id   uuid.UUID       `json:"id"`
		Msg  *models.Message `json:"message"`
		Task T               `json:"task"`
	}{
		r.id,
		r.msg,
		*r.task,
	})
}

func (r Request[T]) SaveToDisk() error {

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
