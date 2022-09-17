package botatobot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
)

type Request struct {
	Task   Txt2img         `json:"task"`
	Id     uuid.UUID       `json:"id"`
	Msg    *models.Message `json:"msg"`
	Output []byte          `json:"-"`
	Err    error           `json:"error,omitempty"`
	Env    string          `json:"env"`
}

func (r Request) GetIdentifier() uuid.UUID {
	return r.Id
}

func (r Request) GetMessage() *models.Message {
	return r.Msg
}

func (r *Request) Launch() {
	r.Output, r.Err = r.Task.Execute(r.Env)
}

func (r Request) Result() ([]byte, error) {
	return r.Output, r.Err
}

func (r Request) String() string {
	return fmt.Sprintf("request %s with parameters: %s, running %s", r.Id, &r.Task, r.Env)
}

func (r Request) SaveToDisk() error {

	err := os.MkdirAll(OUTPUT_PATH, 0755)

	if err != nil {
		return fmt.Errorf("failed to create output directory: %s", err)
	}

	if r.Err != nil {
		return fmt.Errorf("failed to get result: %s", err)
	}

	imgFilePath := filepath.Join(OUTPUT_PATH, fmt.Sprintf("%s.png", r.Id))

	err = os.WriteFile(imgFilePath, r.Output, 0644)

	if err != nil {
		return fmt.Errorf("can't write image to disc: %s", err)
	}

	content, err := json.Marshal(r)

	if err != nil {
		return fmt.Errorf("failed to serialize job parameters: %s", err)
	}

	jsonFilePath := filepath.Join(OUTPUT_PATH, fmt.Sprintf("%s.json", r.Id))

	err = os.WriteFile(jsonFilePath, content, 0644)

	if err != nil {
		return fmt.Errorf("error writing metadata: %s", err)
	}
	return nil
}
