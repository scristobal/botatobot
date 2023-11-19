package pkg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type imageGeneratorKeyTypeWrapper string

const (
	ImageGeneratorKey = imageGeneratorKeyTypeWrapper("replicate")
)

type imageGenerator struct {
	client *http.Client
}

func NewImageGenerator() *imageGenerator {
	return &imageGenerator{
		client: &http.Client{},
	}
}

func (g *imageGenerator) GenerateImageFromPrompt(prompt string) ([]byte, error) {

	input := modelInput{
		Prompt:              prompt,
		Seed:                rand.Intn(1_000_00),
		Num_inference_steps: 50,
		Guidance_scale:      7.5,
	}

	if MODEL_URL == "" {
		return g.generateRemote(input)
	} else {
		return g.generateLocal(input)
	}

}

type modelInput struct {
	Prompt              string  `json:"prompt"`
	Seed                int     `json:"seed,omitempty"`
	Num_inference_steps int     `json:"num_inference_steps,omitempty"`
	Guidance_scale      float32 `json:"guidance_scale,omitempty"`
}

type postResponse struct {
	Urls struct {
		Get string `json:"get"`
	} `json:"urls"`
}

type getResponse struct {
	Output []string `json:"output"`
	Error  string   `json:"error"`
}

func (g *imageGenerator) generateLocal(modelInput modelInput) ([]byte, error) {

	input, err := json.Marshal(modelInput)

	if err != nil {
		return []byte{}, fmt.Errorf("fail to serialize job parameters: %v", err)
	}

	res, err := http.Post(MODEL_URL, "application/json", strings.NewReader(fmt.Sprintf(`{"input": %s}`, input)))

	if err != nil {
		return []byte{}, fmt.Errorf("failed to run the model: %s", err)

	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return []byte{}, fmt.Errorf("can't read model response: %s", err)

	}

	type apiResponse struct {
		Status string   `json:"status"`
		Output []string `json:"output"` // (base64) data URLs
	}

	response := apiResponse{}

	json.Unmarshal(body, &response)

	var output string
	if len(response.Output) > 0 {
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

func (g *imageGenerator) generateRemote(modelInput modelInput) ([]byte, error) {

	// 1st request to launch job
	response, err := g.postJob(modelInput)

	if err != nil {
		// TODO: maybe the api needs more time, try again later
		return []byte{}, fmt.Errorf("failed to post job: %s", err)
	}

	// 2nd request to get job result, eg. output urls
	get_url := response.Urls.Get

	output_urls, err := g.getResponse(get_url)

	if err != nil {
		// TODO: maybe the api needs more time, try again later
		return []byte{}, fmt.Errorf("failed to get job response: %s", err)
	}

	// 3rd request to get image(s)

	// TODO: add support for multiple images
	output_url := output_urls.Output[0]

	data, err := g.getImageData(output_url)

	if err != nil {
		// TODO: maybe the api needs more time, try again later
		return []byte{}, fmt.Errorf("failed to get image data: %s", err)
	}

	return data, nil
}

func (g *imageGenerator) postJob(modelInput modelInput) (postResponse, error) {

	input, err := json.Marshal(modelInput)

	if err != nil {
		return postResponse{}, fmt.Errorf("fail to serialize job parameters: %v", err)
	}

	reqBody := strings.NewReader(fmt.Sprintf(`{"version": "%s", "input": %s}`, REPLICATE_VERSION, input))

	req, err := http.NewRequest("POST", REPLICATE_URL, reqBody)

	if err != nil {
		return postResponse{}, fmt.Errorf("fail to create request: %v", err)
	}

	req.Header.Add("Content-Type", "application/json")

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", REPLICATE_TOKEN))

	res, err := g.client.Do(req)

	if err != nil {
		return postResponse{}, fmt.Errorf("failed to run the model: %s", err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return postResponse{}, fmt.Errorf("can't read model response: %s", err)
	}

	var response postResponse

	json.Unmarshal(body, &response)

	if response.Urls.Get == "" {
		return postResponse{}, fmt.Errorf("can't decode model response: %s", err)
	}

	return response, nil
}

func (g *imageGenerator) getResponse(url string) (getResponse, error) {

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return getResponse{}, fmt.Errorf("fail to create request: %v", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", REPLICATE_TOKEN))

	time.Sleep(5 * time.Second)

	res, err := g.client.Do(req)

	if err != nil {
		return getResponse{}, fmt.Errorf("failed to run the model: %s", err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return getResponse{}, fmt.Errorf("can't read model response: %s", err)
	}

	var resp getResponse

	json.Unmarshal(body, &resp)

	if resp.Error != "" {
		return getResponse{}, fmt.Errorf("problem running the model: %s", resp.Error)
	}

	if len(resp.Output) == -1 {
		return getResponse{}, fmt.Errorf("no output in response")
	}

	if len(resp.Output) == 0 {
		return getResponse{}, fmt.Errorf("empty output in response")
	}

	return resp, nil

}

func (g *imageGenerator) getImageData(url string) ([]byte, error) {

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return []byte{}, fmt.Errorf("fail to create request: %v", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", REPLICATE_TOKEN))

	res, err := g.client.Do(req)

	if err != nil {
		return []byte{}, fmt.Errorf("failed to run the model: %s", err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return []byte{}, fmt.Errorf("can't read model response: %s", err)
	}

	return body, nil
}
