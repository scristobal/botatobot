package worker

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/exp/utf8string"
)

type Txt2img struct {
	Prompt              string  `json:"prompt"`
	Seed                int     `json:"seed,omitempty"`
	Num_inference_steps int     `json:"num_inference_steps,omitempty"`
	Guidance_scale      float32 `json:"guidance_scale,omitempty"`
	Output              []byte  `json:"-"` // not serialized
	Error               error   `json:"error,omitempty"`
	Runner              string
}

type apiResponse struct {
	Status string   `json:"status"`
	Output []string `json:"output"` // (base64) data URLs
}

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

func removeCommands(m string) string {

	words := strings.Split(m, " ")

	for _, w := range words {
		if len(w) > 0 && w[0] == byte('/') {
			m = strings.ReplaceAll(m, w, "")
		}
	}
	return m
}

func removeMentions(m string) string {
	words := strings.Split(m, " ")

	for _, w := range words {
		if len(w) > 0 && w[0] == byte('@') {
			m = strings.ReplaceAll(m, w, "")
		}
	}
	return m
}

func clean(m string) string {
	m = removeMentions(m)

	m = removeCommands(m)

	m = removeSubstrings(m, []string{"\n", "\r", "\t", "\"", "'", ",", "!", "?"})

	m = strings.TrimSpace(m)

	// removes consecutive spaces
	reg := regexp.MustCompile(`\s+`)
	m = reg.ReplaceAllString(m, " ")

	return m
}

func FromString(s string) ([]Txt2img, error) {

	s = clean(s)
	ok := validate(s)

	if !ok {
		return []Txt2img{}, fmt.Errorf("invalid characters in prompt")
	}

	hasSeed := false

	params := Txt2img{
		Num_inference_steps: 50,
		Guidance_scale:      7.5,
		Runner:              "remote",
	}

	words := strings.Split(s, " ")

	for _, word := range words {
		if word[0] == byte('&') {
			split := strings.Split(word, "_")

			if len(split) < 2 {
				return []Txt2img{}, fmt.Errorf("invalid parameter, format should be :param_value")
			}

			key := split[0]
			value := split[1]

			switch key {
			case "&seed":
				seed, err := strconv.Atoi(value)
				if err != nil {
					return []Txt2img{}, fmt.Errorf("invalid seed, should be a number &seed_1234")
				}
				params.Seed = seed
				hasSeed = true
			case "&steps":
				steps, err := strconv.Atoi(value)
				if err != nil {
					return []Txt2img{}, fmt.Errorf("invalid number of inference steps, should be a number &steps_50")
				}

				if steps > 100 || steps < 1 {
					return []Txt2img{}, fmt.Errorf("invalid number of inference steps, should be between 1 and 100 &steps_50")
				}

				params.Num_inference_steps = steps
			case "&guidance":
				guidance, err := strconv.ParseFloat(value, 32)
				if err != nil {
					return []Txt2img{}, fmt.Errorf("invalid guidance scale, should be a rational number &guidance_7.5")
				}
				fmt.Println("guidance", guidance)
				if guidance > 20 || guidance < 1 {
					return []Txt2img{}, fmt.Errorf("invalid guidance scale, should be between 1 and 20 &guidance_7.5")

				}
				params.Guidance_scale = float32(guidance)

			default:
				return []Txt2img{}, fmt.Errorf("invalid parameter, format should be :param_value, allowed parameters are &seed_, &steps_, and &guidance_")
			}

			s = strings.ReplaceAll(s, word, "")
		}
	}

	if len(s) < 10 {
		return []Txt2img{}, fmt.Errorf("prompt too short, should be at least 10 characters")
	}

	params.Prompt = strings.TrimSpace(s)

	if hasSeed {
		return []Txt2img{params}, nil
	}

	jobs := make([]Txt2img, 4)

	for i := 0; i < len(jobs); i++ {

		seed := rand.Intn(1_000_00)

		job := Txt2img{
			Seed:                seed,
			Prompt:              params.Prompt,
			Num_inference_steps: params.Num_inference_steps,
			Guidance_scale:      params.Guidance_scale,
		}

		jobs[i] = job

	}

	return jobs, nil
}

func (j *Txt2img) Run() {

	if j.Runner == "remote" {
		j.Output, j.Error = remoteRunner(j)
	} else {
		j.Output, j.Error = localRunner(j)
	}
}

func (j Txt2img) Result() ([]byte, error) {
	return j.Output, j.Error
}

func (j Txt2img) String() string {

	res := fmt.Sprintf("%s &seed_%d &steps_%d &guidance_%1.f", j.Prompt, j.Seed, j.Num_inference_steps, j.Guidance_scale)

	res = strings.TrimSpace(res)

	return res
}
