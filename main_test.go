package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jarcoal/httpmock"

	"github.com/stretchr/testify/assert"
)

func TestRewriteRouteOk(t *testing.T) {
	limitPerDay = 1000
	rst := getRestyClient()
	httpmock.ActivateNonDefault(rst.GetClient())
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		apiEndpoint,
		httpmock.NewStringResponder(200, `{
			"id": "chatcmpl-8yiKsPfAD4a9d0x1QBLh4kaCgg6lm",
			"object": "chat.completion",
			"created": 1709480798,
			"model": "gpt-3.5-turbo-0125",
			"choices": [
			  {
				"index": 0,
				"message": {
				  "role": "assistant",
				  "content": "You are a witty comedian known for sharing corny dad jokes."
				},
				"logprobs": null,
				"finish_reason": "stop"
			  }
			],
			"usage": {
			  "prompt_tokens": 26,
			  "completion_tokens": 13,
			  "total_tokens": 39
			},
			"system_fingerprint": "fp_2b778c6b35"
		  }`),
	)

	router := setupRouter()

	w := httptest.NewRecorder()

	values := map[string]string{"text": "example text"}

	jsonValue, _ := json.Marshal(values)

	req, _ := http.NewRequest("POST", "/rewrite", bytes.NewBuffer(jsonValue))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "You are a witty comedian known for sharing corny dad jokes.", w.Body.String())
}

func TestContinueRouteOpenaiError(t *testing.T) {
	limitPerDay = 1000
	rst := getRestyClient()
	httpmock.ActivateNonDefault(rst.GetClient())
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		apiEndpoint,
		httpmock.NewStringResponder(200, `{
			"error": {
			  "message": "Incorrect API key provided: k-tuJEjR**************************************iVcJ. You can find your API key at https://platform.openai.com/account/api-keys.",
			  "type": "invalid_request_error",
			  "param": null,
			  "code": "invalid_api_key"
			}
		  }`),
	)

	router := setupRouter()

	w := httptest.NewRecorder()

	values := map[string]string{"text": "example text"}

	jsonValue, _ := json.Marshal(values)

	req, _ := http.NewRequest("POST", "/continue", bytes.NewBuffer(jsonValue))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "Service Unavailable")
}

func TestShortenRouteLimit(t *testing.T) {
	limitPerDay = 1
	timestamps = nil
	// os.Remove(os.TempDir() + filename)
	rst := getRestyClient()
	httpmock.ActivateNonDefault(rst.GetClient())
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		apiEndpoint,
		httpmock.NewStringResponder(200, `{
			"id": "chatcmpl-8yiKsPfAD4a9d0x1QBLh4kaCgg6lm",
			"object": "chat.completion",
			"created": 1709480798,
			"model": "gpt-3.5-turbo-0125",
			"choices": [
			  {
				"index": 0,
				"message": {
				  "role": "assistant",
				  "content": "You are a witty comedian known for sharing corny dad jokes."
				},
				"logprobs": null,
				"finish_reason": "stop"
			  }
			],
			"usage": {
			  "prompt_tokens": 26,
			  "completion_tokens": 13,
			  "total_tokens": 39
			},
			"system_fingerprint": "fp_2b778c6b35"
		  }`),
	)

	router := setupRouter()

	values := map[string]string{"text": "example text"}
	jsonValue, _ := json.Marshal(values)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/shorten", bytes.NewBuffer(jsonValue))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/shorten", bytes.NewBuffer(jsonValue))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}
