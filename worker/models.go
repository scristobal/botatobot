package worker

type Job struct {
	Id     string
	ChatId int
	User   string
	UserId int
	MsgId  int
	Prompt string
	Type   string
}
