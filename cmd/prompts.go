package cmd

import (
	"fmt"
	"regexp"
	"scristobal/botatobot/cfg"
	"strconv"
	"strings"

	"golang.org/x/exp/utf8string"
)

func validate(prompt string) bool {

	ok := utf8string.NewString(prompt).IsASCII()

	if !ok {
		return false
	}

	re := regexp.MustCompile(`^[\w\d\s-:_.|&]*$`)

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

	msg = removeSubstrings(msg, []string{"\n", "\r", "\t", "\"", "'", ",", ".", "!", "?"})

	msg = strings.TrimSpace(msg)

	// removes consecutive spaces
	reg := regexp.MustCompile(`\s+`)
	msg = reg.ReplaceAllString(msg, " ")

	return msg
}

type Params struct {
	Prompt              string `json:"prompt"`
	Seed                *int   `json:"seed,omitempty"`
	Num_inference_steps *int   `json:"num_inference_steps,omitempty"`
	Guidance_scale      *int   `json:"guidance_scale,omitempty"`
}

func (p Params) String() string {
	res := p.Prompt

	if p.Seed != nil {
		res += fmt.Sprintf(" &seed_%d", *p.Seed)
	}

	if p.Num_inference_steps != nil {
		res += fmt.Sprintf(" &steps_%d", *p.Num_inference_steps)
	}

	if p.Guidance_scale != nil {
		res += fmt.Sprintf(" &guidance_%d", *p.Guidance_scale)
	}

	return res
}

func GetParams(msg string) (Params, error) {

	msg = clean(msg)

	ok := validate(msg)

	if !ok {
		return Params{}, fmt.Errorf("invalid characters in prompt")
	}

	words := strings.Split(msg, " ")
	var input Params

	for _, word := range words {
		if word[0] == byte('&') {
			split := strings.Split(word, "_")

			if len(split) < 2 {
				return Params{}, fmt.Errorf("invalid parameter, format should be :param_value")
			}

			key := split[0]
			value := split[1]

			switch key {
			case "&seed":
				seed, err := strconv.Atoi(value)
				if err != nil {
					return Params{}, fmt.Errorf("invalid seed, should be a number &seed_1234")
				}
				input.Seed = &seed
			case "&steps":
				steps, err := strconv.Atoi(value)
				if err != nil {
					return Params{}, fmt.Errorf("invalid number of inference steps, should be a number &steps_50")
				}

				if steps > 100 || steps < 1 {
					return Params{}, fmt.Errorf("invalid number of inference steps, should be between 1 and 100 &steps_50")
				}
				input.Num_inference_steps = &steps
			case "&guidance":
				guidance, err := strconv.Atoi(value)
				if err != nil {
					return Params{}, fmt.Errorf("invalid guidance scale, should be a number &guidance_100")
				}
				if guidance > 20 || guidance < 1 {
					return Params{}, fmt.Errorf("invalid guidance scale, should be between 1 and 20 &guidance_100")
				}
				input.Guidance_scale = &guidance

			default:
				return Params{}, fmt.Errorf("invalid parameter, format should be :param_value, allowed parameters are &seed_, &steps_, and &guidance_")
			}

			msg = strings.ReplaceAll(msg, word, "")
		}
	}

	if len(msg) < 10 {
		return Params{}, fmt.Errorf("prompt too short, should be at least 10 characters")
	}

	input.Prompt = strings.TrimSpace(msg)

	fmt.Println("prompt:", input.Prompt, "--seed:", input.Seed, "--steps:", input.Num_inference_steps, "--guidance:", input.Guidance_scale)

	return input, nil
}
