package tasks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"scristobal/botatobot/cfg"
	"strings"
	"time"
)

type ReplicateTxt2Img struct{ Txt2img }

func (j *ReplicateTxt2Img) Run() {

	input, err := json.Marshal(j)

	if err != nil {
		j.Error = fmt.Errorf("fail to serialize job parameters: %v", err)
		return
	}

	client := &http.Client{}

	// 1st request to launch job

	reqBody := strings.NewReader(fmt.Sprintf(`{"version": "a9758cbfbd5f3c2094457d996681af52552901775aa2d6dd0b17fd15df959bef", "input": %s}`, input))

	req, err := http.NewRequest("POST", cfg.MODEL_URL, reqBody)

	if err != nil {
		j.Error = fmt.Errorf("fail to create request: %v", err)
		return
	}

	req.Header.Add("Content-Type", "application/json")

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", cfg.TOKEN))

	res, err := client.Do(req)

	if err != nil {
		j.Error = fmt.Errorf("failed to run the model: %s", err)
		return
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	fmt.Println("------ 1st call results ------")
	fmt.Println("req", req)
	fmt.Println("res", res)
	fmt.Println("body", string(body))

	if err != nil {
		j.Error = fmt.Errorf("can't read model response: %s", err)
		return
	}

	type apiResponse struct {
		Urls struct {
			Get string `json:"get"`
		} `json:"urls"`
	}

	var response apiResponse

	json.Unmarshal(body, &response)

	if response.Urls.Get == "" {
		j.Error = fmt.Errorf("can't decode model response: %s", err)
		return
	}

	// 2nd request to get job result

	req, err = http.NewRequest("GET", response.Urls.Get, nil)

	if err != nil {
		j.Error = fmt.Errorf("fail to create request: %v", err)
		return
	}

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", cfg.TOKEN))

	time.Sleep(5 * time.Second)

	res, err = client.Do(req)

	if err != nil {
		j.Error = fmt.Errorf("failed to run the model: %s", err)
		return
	}

	defer res.Body.Close()

	body, err = io.ReadAll(res.Body)

	fmt.Println("------ 2nd call results ------")
	fmt.Println("req", req)
	fmt.Println("res", res)
	fmt.Println("body", string(body))

	if err != nil {
		j.Error = fmt.Errorf("can't read model response: %s", err)
		return
	}

	type getResponse struct {
		Output []string `json:"output"`
	}

	var resp getResponse

	json.Unmarshal(body, &resp)

	if len(resp.Output) == 0 {
		j.Error = fmt.Errorf("can't decode model response: %s", err)
		return
	}

	// 3rd request to get image

	req, err = http.NewRequest("GET", resp.Output[0], nil)

	if err != nil {
		j.Error = fmt.Errorf("fail to create request: %v", err)
		return
	}

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", cfg.TOKEN))

	res, err = client.Do(req)

	if err != nil {
		j.Error = fmt.Errorf("failed to run the model: %s", err)
		return
	}

	defer res.Body.Close()

	body, err = io.ReadAll(res.Body)

	//fmt.Println("------ 3rd call results ------")
	//fmt.Println("req", req)
	//fmt.Println("res", res)
	//fmt.Println("body", string(body))

	if err != nil {
		j.Error = fmt.Errorf("can't read model response: %s", err)
		return
	}

	j.Output = body

}
