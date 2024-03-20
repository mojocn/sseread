package sseread

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// https://developers.cloudflare.com/workers-ai/models/zephyr-7b-beta-awq/#using-streaming
// CfTextGenerationResponse represents the response structure for text generation from the Cloudflare AI API.
type CfTextGenerationResponse struct {
	Response string `json:"response"`
	P        string `json:"p"`
}

// CfTextGenerationMsg represents a message for text generation in Cloudflare AI.
type CfTextGenerationMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CfTextGenerationArg represents the arguments for text generation in Cloudflare AI.
type CfTextGenerationArg struct {
	Stream   bool                  `json:"stream,omitempty"`
	Messages []CfTextGenerationMsg `json:"messages,omitempty"`
}

// body returns the request body as an io.ReadCloser and any encoding error encountered.
func (c *CfTextGenerationArg) body() (io.ReadCloser, error) {
	buff := bytes.NewBuffer(nil)
	err := json.NewEncoder(buff).Encode(c)
	return io.NopCloser(buff), err
}

// CloudflareAI represents the configuration for accessing the Cloudflare AI service.
type CloudflareAI struct {
	AccountID string // AccountID is the identifier for the Cloudflare account.
	APIToken  string // APIToken is the authentication token for accessing the Cloudflare AI service.
}

var httpClient = &http.Client{}

// modelsTextGeneration is a slice of strings that represents a list of models used for text generation.
// Each string in the slice corresponds to a specific model.
// The models are stored in the Cloudflare AI service and can be accessed using the provided URLs.
// The list is divided into multiple pages, with each page containing a set of models.
// To access a specific model, you can refer to its corresponding URL.
var modelsTextGeneration = []string{
	//https://dash.cloudflare.com/0a76b889e644c012524110042e6f197e/ai/workers-ai
	//page 1
	"@cf/meta/llama-2-7b-chat-fp16",
	"@cf/mistral/mistral-7b-instruct-v0.1",
	"@cf/meta/llama-2-7b-chat-int8",
	"@cf/qwen/qwen1.5-0.5b-chat",
	"@hf/thebloke/llamaguard-7b-awq",
	"@hf/thebloke/neural-chat-7b-v3-1-awq",
	"@cf/deepseek-ai/deepseek-math-7b-base",
	"@cf/tinyllama/tinyllama-1.1b-chat-v1.0",
	"@hf/thebloke/orca-2-13b-awq",
	"@hf/thebloke/codellama-7b-instruct-awq",
	//page 2
	"@cf/thebloke/discolm-german-7b-v1-awq",
	"@hf/thebloke/mistral-7b-instruct-v0.1-awq",
	"@hf/thebloke/openchat_3.5-awq",
	"@cf/qwen/qwen1.5-7b-chat-awq",
	"@hf/thebloke/llama-2-13b-chat-awq",
	"@hf/thebloke/deepseek-coder-6.7b-base-awq",
	"@hf/thebloke/openhermes-2.5-mistral-7b-awq",
	"@hf/thebloke/deepseek-coder-6.7b-instruct-awq",
	"@cf/deepseek-ai/deepseek-math-7b-instruct",
	"@cf/tiiuae/falcon-7b-instruct",
	//page 3
	"@hf/thebloke/zephyr-7b-beta-awq",
	"@cf/qwen/qwen1.5-1.8b-chat",
	"@cf/defog/sqlcoder-7b-2",
	"@cf/microsoft/phi-2",
	"@cf/qwen/qwen1.5-14b-chat-awq",
	"@cf/openchat/openchat-3.5-0106",
}

func (c *CloudflareAI) modelCheck(model string) error {
	for _, v := range modelsTextGeneration {
		if v == model {
			return nil
		}
	}
	return errors.New("model not found: " + model)
}

// Do executes the Cloudflare AI model with the specified model and arguments.
// It returns the HTTP response and an error, if any.
func (c *CloudflareAI) Do(model string, arg *CfTextGenerationArg) (*http.Response, error) {
	if c.AccountID == "" || c.APIToken == "" {
		return nil, errors.New("CF_ACCOUNT_ID and CF_API_TOKEN environment variables are required")
	}

	if err := c.modelCheck(model); err != nil {
		return nil, err
	}

	body, err := arg.body()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.cloudflare.com/client/v4/accounts/"+c.AccountID+"/ai/run/"+model, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIToken)
	return httpClient.Do(req)
}
