package pkg

import (
	"encoding/json"
	"fmt"
	"io"
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

func (g *imageGenerator) GenerateImageFromPrompt(prompt string) ([]string, error) {

	modelInput := modelInput{
		Prompt:               prompt,
		NumOutputs:           4,
		DisableSafetyChecker: true,
	}

	// 1st request to launch job
	response, err := g.postJob(modelInput)

	if err != nil {
		// TODO: maybe the api needs more time, try again later
		return []string{}, fmt.Errorf("failed to post job: %s", err)
	}

	// 2nd request to get job result, eg. output urls
	getUrl := response.Urls.Get

	tryCount := 0

	responsesUrls, err := g.getResponse(getUrl)

	for err != nil && tryCount < 5 {
		// maybe the api needs more time, try again later
		time.Sleep(1 * time.Second)
		responsesUrls, err = g.getResponse(getUrl)
	}

	if err != nil {
		return []string{}, fmt.Errorf("failed to get job response: %s", err)
	}

	return responsesUrls.Output, nil
}

type modelInput struct {
	Prompt               string `json:"prompt"`
	NumOutputs           int32  `json:"num_outputs"`
	DisableSafetyChecker bool   `json:"disable_safety_checker"`
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
