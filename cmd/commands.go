package cmd

type Command int8

// commands
const (
	Help Command = iota
	Generate
	Status
)

var commands = []Command{Help, Generate, Status}

func (c Command) String() string {
	switch c {
	case Help:
		return "/help"
	case Generate:
		return "/generate"

	case Status:
		return "/status"
	}
	return "unknown"
}
