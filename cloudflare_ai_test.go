package sseread

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
)

// https://developers.cloudflare.com/workers-ai/models/zephyr-7b-beta-awq/#using-streaming
type llamaMsg struct {
	Response string `json:"response"`
	P        string `json:"p"`
}

// TestReadFromCloudflareLama2 is a test function for the ReadCh function in the sseread package.
// It sends a POST request to the Cloudflare API and reads the response body as Server-Sent Events.
// For each event, it parses the JSON object from the event data and appends the response to the fulltext string.
// If an error occurs during the POST request, the reading of the events, or the JSON unmarshalling, it fails the test.
func TestReadFromCloudflareLama2(t *testing.T) {
	// Retrieve the account ID and API token from the environment variables
	accountID := os.Getenv("CF_ACCOUNT_ID")
	apiToken := os.Getenv("CF_API_TOKEN")
	if accountID == "" || apiToken == "" {
		t.Fatal("CF_ACCOUNT_ID and CF_API_TOKEN environment variables are required")
	}
	// Create a buffer with the request body
	buff := bytes.NewBufferString(`{ "stream":true,"messages": [{ "role": "system", "content": "You are a friendly assistant" }, { "role": "user", "content": "Why is pizza so good" }]}`)

	// Create a new POST request to the Cloudflare API
	req, err := http.NewRequest("POST", "https://api.cloudflare.com/client/v4/accounts/"+accountID+"/ai/run/@cf/meta/llama-2-7b-chat-int8", buff)
	if err != nil {
		t.Fatal(err)
	}

	// Set the Authorization header with the API token
	req.Header.Set("Authorization", "Bearer "+apiToken)

	// Send the POST request
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	// Ensure the response body is closed after the function returns
	defer response.Body.Close()

	// Check the response status code
	if response.StatusCode != http.StatusOK {
		all, err := io.ReadAll(response.Body)
		if err != nil {
			t.Error(err)
		}
		t.Log(string(all))
		return
	}

	// Read the response body as Server-Sent Events
	channel, err := ReadCh(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	// Initialize an empty string to store the full text of the responses
	fulltext := ""

	// Iterate over the events from the channel
	for event := range channel {
		if event == nil || event.IsSkip() {
			continue
		}

		// Parse the JSON object from the event data
		e := new(llamaMsg)
		err := json.Unmarshal(event.Data, e)
		if err != nil {
			t.Error(err, string(event.Data))
		} else {
			// Append the response to the fulltext string
			fulltext += e.Response
		}
	}

	// Log the full text of the responses
	t.Log(fulltext)
}