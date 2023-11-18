package botatobot

type Command string

const (
	HelpCmd     Command = "/help"
	GenerateCmd Command = "/generate"
	StatusCmd   Command = "/status"
)
