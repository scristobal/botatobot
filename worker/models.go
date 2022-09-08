package worker

import "sync"

type Job struct {
	Id     string
	ChatId int
	User   string
	UserId int
	MsgId  int
	Prompt string
	Type   string
}

type currentJob struct {
	job *Job
	mut sync.RWMutex
}
