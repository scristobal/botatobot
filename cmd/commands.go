package cmd

type Command string

// commands
const (
	Help     Command = "/help"
	Generate Command = "/generate"
	Status   Command = "/status"
)

var commands = []Command{Help, Generate, Status}

func (c Command) String() string {
	switch c {
	case Help:
		return "Get Help"
	case Generate:
		return "Generate a text from a prompt"

	case Status:
		return "Generate a text from a prompt"
	}

	return "Unknown command"
}
