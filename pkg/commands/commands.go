package commands

type Command string

const (
	Help     Command = "/help"
	Generate Command = "/generate"
	Status   Command = "/status"
)
