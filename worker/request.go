package worker

import (
	"fmt"
	"scristobal/botatobot/tasks"

	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
)

type Request struct {
	tasks.Txt2img
	Id  uuid.UUID       `json:"id"`
	Msg *models.Message `json:"message"`
}

func New(m models.Message) ([]Request, error) {

	jobs, err := tasks.FromString(m.Text)

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}

	var requests []Request

	for _, job := range jobs {
		requests = append(requests, Request{job, uuid.New(), &m})
	}

	return requests, nil
}
