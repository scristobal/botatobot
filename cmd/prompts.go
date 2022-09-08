package cmd

import (
	"fmt"
	"regexp"
	"scristobal/botatobot/cfg"
	"strings"

	"golang.org/x/exp/utf8string"
)

func validate(prompt string) bool {

	ok := utf8string.NewString(prompt).IsASCII()

	if !ok {
		return false
	}

	re := regexp.MustCompile(`^[\w\d\s-:.]*$`)

	return re.MatchString(prompt) && len(prompt) > 0
}

func removeSubstrings(s string, b []string) string {
	for _, c := range b {
		s = strings.ReplaceAll(s, c, "")
	}
	return s
}

func removeCommands(msg string) string {
	for _, c := range commands {
		msg = strings.ReplaceAll(msg, string(c), "")
	}
	return msg
}

func removeBotName(msg string) string {
	return strings.ReplaceAll(msg, cfg.BOT_USERNAME, "")
}

func clean(msg string) string {
	msg = removeBotName(msg)

	msg = removeCommands(msg)

	msg = removeSubstrings(msg, []string{"\n", "\r", "\t", "\"", "'", ",", ".", "!", "?", "_"})

	msg = strings.TrimSpace(msg)

	// removes consecutive spaces
	reg := regexp.MustCompile(`\s+`)
	msg = reg.ReplaceAllString(msg, " ")

	return msg
}

func GetPrompt(msg string) (string, error) {

	prompt := clean(msg)

	ok := validate(prompt)

	if !ok {
		return "", fmt.Errorf("invalid characters in prompt")
	}

	return prompt, nil
}
