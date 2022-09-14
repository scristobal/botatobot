package worker

import (
	"fmt"
	"scristobal/botatobot/tasks"

	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
)

type Request struct {
	t   *tasks.Txt2img
	id  uuid.UUID       //`json:"id"`
	msg *models.Message //`json:"message"`
}

func New(m models.Message) ([]Request, error) {

	jobs, err := tasks.FromString(m.Text)

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}

	var requests []Request

	for _, job := range jobs {
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
