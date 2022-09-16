package scheduler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"scristobal/botatobot/config"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/utf8string"
)

type Txt2img struct {
	Prompt              string  `json:"prompt"`
	Seed                int     `json:"seed,omitempty"`
	Num_inference_steps int     `json:"num_inference_steps,omitempty"`
	Guidance_scale      float32 `json:"guidance_scale,omitempty"`
	Output              []byte  `json:"-"` // not serialized
	Error               error   `json:"error,omitempty"`
	Env                 string
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

func removeConsecutiveSpaces(s string) string {
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(s, " ")
}

func clean(m string) string {

	m = removeConsecutiveSpaces(m)

	m = removeMentions(m)

	m = removeCommands(m)

	m = removeSubstrings(m, []string{"\n", "\r", "\t", "\"", "'", ",", "!", "?"})

	m = strings.TrimSpace(m)

	return m
}

func getParams(s string) (map[string]string, error) {

	result := make(map[string]string)

	words := strings.Split(s, " ")

	for _, word := range words {
		if len(word) > 0 && word[0] == byte('&') {
			split := strings.Split(word, "_")

			if len(split) != 2 {
				return result, fmt.Errorf("invalid parameter %s, format should be :param_value", word)
			}

			result[split[0]] = split[1]
		}
	}

	return result, nil
}

func getPrompt(s string) (string, error) {

	s = clean(s)
	ok := validate(s)

	if !ok {
		return "", fmt.Errorf("invalid characters in prompt")
	}

	words := strings.Split(s, " ")

	var result []string

	for _, w := range words {
		if len(w) > 0 && w[0] != byte('&') {
			result = append(result, w)
		}
	}

	prompt := strings.Join(result, " ")

	if len(prompt) < 10 {
		return "", fmt.Errorf("prompt too short, should be at least 10 characters")
	}

	return prompt, nil

}

func buildConfig(prompt string, params map[string]string) (Txt2img, error) {

	config := Txt2img{
		Prompt:              prompt,
		Num_inference_steps: 50,
		Guidance_scale:      7.5,
		Env:                 "remote",
	}

	for key, value := range params {
		switch key {
		case "&seed":
			seed, err := strconv.Atoi(value)
			if err != nil {
				return config, fmt.Errorf("invalid seed, should be a number &seed_1234")
			}
			config.Seed = seed

		case "&steps":
			steps, err := strconv.Atoi(value)
			if err != nil {
				return config, fmt.Errorf("invalid number of inference steps, should be a number &steps_50")
			}

			if steps > 100 || steps < 1 {
				return config, fmt.Errorf("invalid number of inference steps, should be between 1 and 100 &steps_50")
			}

			config.Num_inference_steps = steps
		case "&guidance":
			guidance, err := strconv.ParseFloat(value, 32)
			if err != nil {
				return config, fmt.Errorf("invalid guidance scale, should be a rational number &guidance_7.5")
			}
			if guidance > 20 || guidance < 1 {
				return config, fmt.Errorf("invalid guidance scale, should be between 1 and 20 &guidance_7.5")

			}
			config.Guidance_scale = float32(guidance)

		default:
			return config, fmt.Errorf("invalid parameter, format should be :param_value, allowed parameters are &seed_, &steps_, and &guidance_")
		}

	}

	return config, nil

}

func FromString(s string) ([]*Txt2img, error) {

	prompt, err := getPrompt(s)

	if err != nil {
		return nil, fmt.Errorf("invalid prompt: %s", err)
	}

	userParams, err := getParams(s)

	if err != nil {
		return nil, fmt.Errorf("invalid parameters: %s", err)
	}

	params, err := buildConfig(prompt, userParams)

	if err != nil {
		return []*Txt2img{}, fmt.Errorf("invalid parameters, %s", err)
	}

	// special case, if no seed is provided we generate 5 images with different seeds
	_, ok := userParams["&seed"]

	if ok {
		return []*Txt2img{&params}, nil
	}

	jobs := make([]*Txt2img, 4)

	for i := 0; i < len(jobs); i++ {

		seed := rand.Intn(1_000_00)

		job := Txt2img{
			Seed:                seed,
			Prompt:              params.Prompt,
			Num_inference_steps: params.Num_inference_steps,
			Guidance_scale:      params.Guidance_scale,
			Env:                 params.Env,
		}

		jobs[i] = &job

	}

	return jobs, nil
}

func localRunner(j *Txt2img) ([]byte, error) {
	input, err := json.Marshal(j)

	if err != nil {
		return []byte{}, fmt.Errorf("fail to serialize job parameters: %v", err)
	}

	res, err := http.Post(config.MODEL_URL, "application/json", strings.NewReader(fmt.Sprintf(`{"input": %s}`, input)))

	if err != nil {
		return []byte{}, fmt.Errorf("failed to run the model: %s", err)

	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return []byte{}, fmt.Errorf("can't read model response: %s", err)

	}

	response := apiResponse{}

	json.Unmarshal(body, &response)

	var output string
	if len(response.Output) > 0 { // local response from replicate
		output = response.Output[0]

		// remove the data URL prefix
		data := strings.SplitAfter(output, ",")[1]

		decoded, err := base64.StdEncoding.DecodeString(data)

		if err != nil {
			return []byte{}, fmt.Errorf("can't decode model response: %s", err)

		}

		return decoded, nil
	} else {
		return []byte{}, fmt.Errorf("no output in model response")
	}

}

func remoteRunner(j *Txt2img) ([]byte, error) {
	fmt.Println("got job")
	input, err := json.Marshal(j)

	if err != nil {
		return []byte{}, fmt.Errorf("fail to serialize job parameters: %v", err)
	}

	client := &http.Client{}

	// 1st request to launch job

	reqBody := strings.NewReader(fmt.Sprintf(`{"version": "a9758cbfbd5f3c2094457d996681af52552901775aa2d6dd0b17fd15df959bef", "input": %s}`, input))

	req, err := http.NewRequest("POST", config.MODEL_URL, reqBody)

	if err != nil {
		return []byte{}, fmt.Errorf("fail to create request: %v", err)
	}

	req.Header.Add("Content-Type", "application/json")

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", config.TOKEN))

	res, err := client.Do(req)

	if err != nil {
		return []byte{}, fmt.Errorf("failed to run the model: %s", err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	fmt.Println("------ 1st call results ------")
	fmt.Println("req", req)
	fmt.Println("res", res)
	fmt.Println("body", string(body))

	if err != nil {
		return []byte{}, fmt.Errorf("can't read model response: %s", err)
	}

	type apiResponse struct {
		Urls struct {
			Get string `json:"get"`
		} `json:"urls"`
	}

	var response apiResponse

	json.Unmarshal(body, &response)

	if response.Urls.Get == "" {
		return []byte{}, fmt.Errorf("can't decode model response: %s", err)
	}

	// 2nd request to get job result

	req, err = http.NewRequest("GET", response.Urls.Get, nil)

	if err != nil {
		return []byte{}, fmt.Errorf("fail to create request: %v", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", config.TOKEN))

	time.Sleep(5 * time.Second)

	res, err = client.Do(req)

	if err != nil {
		return []byte{}, fmt.Errorf("failed to run the model: %s", err)
	}

	defer res.Body.Close()

	body, err = io.ReadAll(res.Body)

	fmt.Println("------ 2nd call results ------")
	fmt.Println("req", req)
	fmt.Println("res", res)
	fmt.Println("body", string(body))

	if err != nil {
		return []byte{}, fmt.Errorf("can't read model response: %s", err)
	}

	type getResponse struct {
		Output []string `json:"output"`
		Error  string   `json:"error"`
	}

	var resp getResponse

	json.Unmarshal(body, &resp)

	if resp.Error != "" {
		return []byte{}, fmt.Errorf("problem running the model: %s", resp.Error)
	}

	if len(resp.Output) == 0 {
		return []byte{}, fmt.Errorf("empty model response")
	}

	// 3rd request to get image

	req, err = http.NewRequest("GET", resp.Output[0], nil)

	if err != nil {
		return []byte{}, fmt.Errorf("fail to create request: %v", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", config.TOKEN))

	res, err = client.Do(req)

	if err != nil {
		return []byte{}, fmt.Errorf("failed to run the model: %s", err)
	}

	defer res.Body.Close()

	body, err = io.ReadAll(res.Body)

	//fmt.Println("------ 3rd call results ------")
	//fmt.Println("req", req)
	//fmt.Println("res", res)
	//fmt.Println("body", string(body))

	if err != nil {
		return []byte{}, fmt.Errorf("can't read model response: %s", err)
	}

	return body, nil
}

func (j *Txt2img) Launch() {
	if j.Env == "remote" {
		j.Output, j.Error = remoteRunner(j)
	} else {
		j.Output, j.Error = localRunner(j)
	}
}

func (j *Txt2img) Result() ([]byte, error) {
	return j.Output, j.Error
}

func (j *Txt2img) Describe() string {

	res := fmt.Sprintf("%s &seed_%d &steps_%d &guidance_%1.f", j.Prompt, j.Seed, j.Num_inference_steps, j.Guidance_scale)

	res = strings.TrimSpace(res)

	return res
}
