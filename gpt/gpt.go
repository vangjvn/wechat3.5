package gpt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/qingconglaixueit/wechatbot/config"
)

const ModelGpt35Turbo = "gpt-3.5-turbo"

const MaxTokensGpt35Turbo = 4096

const (
	RoleUser      RoleType = "user"
	RoleAssistant RoleType = "assistant"
	RoleSystem    RoleType = "system"
)

// Completions gtp文本模型回复
// curl https://api.openai.com/v1/completions
// -H "Content-Type: application/json"
// -H "Authorization: Bearer your chatGPT key"
// -d '{"model": "text-davinci-003", "prompt": "give me good song", "temperature": 0, "max_tokens": 7}'
func Completions(msg string) (string, error) {
	var gptResponseBody *Response
	var resErr error
	for retry := 1; retry <= 3; retry++ {
		if retry > 1 {
			time.Sleep(time.Duration(retry-1) * 100 * time.Millisecond)
		}
		gptResponseBody, resErr = httpRequestCompletions(msg, retry)
		if resErr != nil {
			log.Printf("gpt request(%d) error: %v\n", retry, resErr)
			continue
		}
		//if gptResponseBody.Error.Message == "" {
		//	break
		//}
	}
	if resErr != nil {
		return "", resErr
	}
	var reply string
	if gptResponseBody != nil && len(gptResponseBody.Choices) > 0 {
		reply = gptResponseBody.Choices[0].Message.Content
	}
	return reply, nil
}

func httpRequestCompletions(msg string, runtimes int) (*Response, error) {
	cfg := config.LoadConfig()
	if cfg.ApiKey == "" {
		return nil, errors.New("api key required")
	}
	requestBody := Request{
		Model: ModelGpt35Turbo,
		Messages: []*Message{
			{RoleUser,
				msg},
		},
		Temperature: cfg.Temperature,
		MaxTokens:   cfg.MaxTokens,
	}
	jsonData, err := json.Marshal(requestBody)
	fmt.Println("--------------------------------", requestBody)
	//requestData, err := json.Marshal(requestBody)
	//fmt.Println("--------------------------------", requestData)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal requestBody error: %v", err)
	}
	//
	//log.Printf("gpt request(%d) json: %s\n", runtimes, string(requestData))

	url1 := "http://openai.rdrstartup.com"
	url2 := "http://openai.chinardr.com"

	//url1,url2 = url2,url1
	req, err := http.NewRequest(http.MethodPost, url1, bytes.NewBuffer(jsonData))

	if err != nil {
		req, err = http.NewRequest(http.MethodPost, url2, bytes.NewBuffer(jsonData))
	}
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest error: %v", err)
	}

	//req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/completions", bytes.NewBuffer(requestData))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.ApiKey)
	client := &http.Client{Timeout: 20 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client.Do error: %v", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadAll error: %v", err)
	}

	log.Printf("gpt response(%d) json: %s\n", runtimes, string(body))

	gptResponseBody := &Response{}
	err = json.Unmarshal(body, gptResponseBody)
	//err = json.NewDecoder(response.Body).Decode(&gptResponseBody)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal responseBody error: %v", err)
	}
	return gptResponseBody, nil
}
