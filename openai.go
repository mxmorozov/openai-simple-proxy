package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/go-resty/resty/v2"
)

const (
	apiEndpoint = "https://api.openai.com/v1/chat/completions"
)

var _restyClient *resty.Client
var lock = &sync.Mutex{}
var apiKey = os.Getenv("OPENAI_API_KEY")

func getRestyClient() *resty.Client {
	if _restyClient == nil {
		lock.Lock()
		defer lock.Unlock()
		if _restyClient == nil {
			_restyClient = resty.New()
		}
	}

	return _restyClient
}

func chatgptRewrite(text string) (string, error) {
	return chatgptRequest("Rewrite the text", text)
}

func chatgptContinue(text string) (string, error) {
	content, err := chatgptRequest("Continue the text", text)
	if err == nil {
		content = text + " " + content
	}
	return content, err
}

func chatgptShorten(text string) (string, error) {
	return chatgptRequest("Shorten the text", text)
}

func chatgptRequest(query string, text string) (string, error) {

	client := getRestyClient()

	response, err := client.R().
		SetAuthToken(apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"model":      "gpt-3.5-turbo",
			"messages":   []interface{}{map[string]interface{}{"role": "system", "content": query + `: "` + text + `"`}},
			"max_tokens": 1000,
		}).
		Post(apiEndpoint)

	if err != nil {
		log.Fatalf("Error while sending the request: %v", err)
		return "", err
	}

	body := response.Body()

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)

	if err != nil {
		log.Fatalf("Error while decoding JSON response: %v", err)
		return "", err
	}

	_, ok := data["choices"]

	if ok {
		content := data["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
		return content, nil
	}

	_, ok = data["error"]

	if ok {
		errMessage := data["error"].(map[string]interface{})["message"].(string)
		return "", errors.New(errMessage)
	}

	return "", fmt.Errorf("incomprehensible answer %v", data)
}
