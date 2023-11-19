package botatobot

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

type ModelInput struct {
	Prompt              string  `json:"prompt"`
	Seed                int     `json:"seed,omitempty"`
	Num_inference_steps int     `json:"num_inference_steps,omitempty"`
	Guidance_scale      float32 `json:"guidance_scale,omitempty"`
}

func generateInput(prompt string) ModelInput {

	// TODO: sanitize prompt
	return ModelInput{
		Prompt:              prompt,
		Seed:                rand.Intn(1_000_00),
		Num_inference_steps: 50,
		Guidance_scale:      7.5,
	}
}

func GenerateImage(prompt string) ([]byte, error) {

	if MODEL_URL == "" {
		return generateRemote(prompt)
	} else {
		return generateLocal(prompt)
	}

}

func generateLocal(prompt string) ([]byte, error) {

	modelInput := generateInput(prompt)

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

func generateRemote(prompt string) ([]byte, error) {

	modelInput := generateInput(prompt)

	input, err := json.Marshal(modelInput)

	if err != nil {
		return []byte{}, fmt.Errorf("fail to serialize job parameters: %v", err)
	}

	client := &http.Client{}

	// 1st request to launch job

	version := REPLICATE_VERSION

	reqBody := strings.NewReader(fmt.Sprintf(`{"version": "%s", "input": %s}`, version, input))

	req, err := http.NewRequest("POST", REPLICATE_URL, reqBody)

	if err != nil {
		return []byte{}, fmt.Errorf("fail to create request: %v", err)
	}

	req.Header.Add("Content-Type", "application/json")

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", REPLICATE_TOKEN))

	res, err := client.Do(req)

	if err != nil {
		return []byte{}, fmt.Errorf("failed to run the model: %s", err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	/*
		fmt.Println("------ 1st call results ------")
		fmt.Println("req", req)
		fmt.Println("res", res)
		fmt.Println("body", string(body))
	*/

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

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", REPLICATE_TOKEN))

	time.Sleep(5 * time.Second)

	res, err = client.Do(req)

	if err != nil {
		return []byte{}, fmt.Errorf("failed to run the model: %s", err)
	}

	defer res.Body.Close()

	body, err = io.ReadAll(res.Body)

	/*
		fmt.Println("------ 2nd call results ------")
		fmt.Println("req", req)
		fmt.Println("res", res)
		fmt.Println("body", string(body))
	*/

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

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", REPLICATE_TOKEN))

	res, err = client.Do(req)

	if err != nil {
		return []byte{}, fmt.Errorf("failed to run the model: %s", err)
	}

	defer res.Body.Close()

	body, err = io.ReadAll(res.Body)

	/*
		fmt.Println("------ 3rd call results ------")
		fmt.Println("req", req)
		fmt.Println("res", res)
		fmt.Println("body", string(body))
	*/

	if err != nil {
		return []byte{}, fmt.Errorf("can't read model response: %s", err)
	}

	return body, nil
}
