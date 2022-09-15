package jobs

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"scristobal/botatobot/cfg"
	"strconv"
	"strings"

	"golang.org/x/exp/utf8string"
)

type Params struct {
	Prompt              string  `json:"prompt"`
	Seed                int     `json:"seed,omitempty"`
	Num_inference_steps int     `json:"num_inference_steps,omitempty"`
	Guidance_scale      float32 `json:"guidance_scale,omitempty"`
}

type Txt2img struct {
	Id     string
	ChatId int
	User   string
	UserId int
	MsgId  int
	Params Params
	Type   string
}

type modelResponse struct {
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

func removeCommands(msg string) string {

	words := strings.Split(msg, " ")

	for _, w := range words {
		if len(w) > 0 && w[0] == byte('/') {
			msg = strings.ReplaceAll(msg, w, "")
		}
	}
	return msg
}

func removeMentions(msg string) string {
	words := strings.Split(msg, " ")

	for _, w := range words {
		if len(w) > 0 && w[0] == byte('@') {
			msg = strings.ReplaceAll(msg, w, "")
		}
	}
	return msg
}

func clean(msg string) string {
	msg = removeMentions(msg)

	msg = removeCommands(msg)

	msg = removeSubstrings(msg, []string{"\n", "\r", "\t", "\"", "'", ",", "!", "?"})

	msg = strings.TrimSpace(msg)

	// removes consecutive spaces
	reg := regexp.MustCompile(`\s+`)
	msg = reg.ReplaceAllString(msg, " ")

	return msg
}

func (p Params) String() string {
	res := p.Prompt

	res += fmt.Sprintf(" &seed_%d", p.Seed)

	res += fmt.Sprintf(" &steps_%d", p.Num_inference_steps)

	res += fmt.Sprintf(" &guidance_%.1f", p.Guidance_scale)

	res = strings.TrimSpace(res)

	return res
}

func GetParams(msg string) (Params, bool, error) {

	msg = clean(msg)
	ok := validate(msg)

	if !ok {
		return Params{}, false, fmt.Errorf("invalid characters in prompt")
	}

	hasParams := false

	input := Params{
		Prompt:              "",
		Seed:                rand.Intn(1000000),
		Num_inference_steps: 50,
		Guidance_scale:      7.5,
	}

	words := strings.Split(msg, " ")

	for _, word := range words {
		if word[0] == byte('&') {
			split := strings.Split(word, "_")

			if len(split) < 2 {
				return Params{}, hasParams, fmt.Errorf("invalid parameter, format should be :param_value")
			}

			key := split[0]
			value := split[1]

			switch key {
			case "&seed":
				hasParams = true
				seed, err := strconv.Atoi(value)
				if err != nil {
					return Params{}, hasParams, fmt.Errorf("invalid seed, should be a number &seed_1234")
				}
				input.Seed = seed
			case "&steps":
				hasParams = true
				steps, err := strconv.Atoi(value)
				if err != nil {
					return Params{}, hasParams, fmt.Errorf("invalid number of inference steps, should be a number &steps_50")
				}

				if steps > 100 || steps < 1 {
					return Params{}, hasParams, fmt.Errorf("invalid number of inference steps, should be between 1 and 100 &steps_50")
				}
				input.Num_inference_steps = steps
			case "&guidance":
				hasParams = true
				guidance, err := strconv.ParseFloat(value, 32)
				if err != nil {
					return Params{}, hasParams, fmt.Errorf("invalid guidance scale, should be a rational number &guidance_7.5")
				}
				fmt.Println("guidance", guidance)
				if guidance > 20 || guidance < 1 {
					return Params{}, hasParams, fmt.Errorf("invalid guidance scale, should be between 1 and 20 &guidance_7.5")
				}
				input.Guidance_scale = float32(guidance)

			default:
				return Params{}, hasParams, fmt.Errorf("invalid parameter, format should be :param_value, allowed parameters are &seed_, &steps_, and &guidance_")
			}

			msg = strings.ReplaceAll(msg, word, "")
		}
	}

	if len(msg) < 10 {
		return Params{}, hasParams, fmt.Errorf("prompt too short, should be at least 10 characters")
	}

	input.Prompt = strings.TrimSpace(msg)

	fmt.Println("prompt:", input.Prompt, "--seed:", input.Seed, "--steps:", input.Num_inference_steps, "--guidance:", input.Guidance_scale)

	return input, hasParams, nil
}

func (job Txt2img) Run() {

	input, err := json.Marshal(job.Params)

	if err != nil {
		log.Printf("error marshaling input: %v", err)
		return
	}

	res, err := http.Post(cfg.MODEL_URL, "application/json", strings.NewReader(fmt.Sprintf(`{"input": %s}`, input)))

	if err != nil {
		log.Printf("Error job %s while requesting model: %s\n", job.Id, err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		log.Printf("Error job %s while reading model response: %s\n", job.Id, err)
		return
	}

	response := modelResponse{}

	json.Unmarshal(body, &response)

	output := response.Output[0]

	// remove the data URL prefix
	data := strings.SplitAfter(output, ",")[1]

	decoded, err := base64.StdEncoding.DecodeString(data)

	if err != nil {
		log.Printf("Error job %s while decoding model response: %s\n", job.Id, err)
		return
	}

	imgFilePath := filepath.Join(cfg.OUTPUT_PATH, fmt.Sprintf("%s.png", job.Id))

	err = os.WriteFile(imgFilePath, decoded, 0644)

	if err != nil {
		log.Printf("Error job %s while writing image: %s\n", job.Id, err)
		return
	}

	content, err := json.Marshal(job)

	if err != nil {
		log.Printf("Error marshalling job %v", err)
		return
	}

	jsonFilePath := filepath.Join(cfg.OUTPUT_PATH, fmt.Sprintf("%s.json", job.Id))

	err = os.WriteFile(jsonFilePath, content, 0644)

	if err != nil {
		log.Printf("Error writing meta.json of job %s: %v", job.Id, err)
		return
	}
}
