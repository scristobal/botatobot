package worker

import (
	"fmt"
	"scristobal/botatobot/jobs"

	"github.com/google/uuid"
)

type Requester struct {
	ChatId int
	User   string
	UserId int
	MsgId  int
	Msg    string
}

type Request struct {
	Id        uuid.UUID
	Requester Requester
	Jobs      []jobs.Txt2img
}

func New(r Requester) (Request, error) {

	id := uuid.New()

	jobs, err := jobs.FromString(r.Msg)

	if err != nil {
		return Request{}, fmt.Errorf("failed to create request: %s", err)
	}

	return Request{Id: id, Requester: r, Jobs: jobs}, nil
}
