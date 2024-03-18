package sseread

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
)

// TestReadFromCloudflareLama2 is a test function for the ReadCh function in the sseread package.
// It sends a POST request to the Cloudflare API and reads the response body as Server-Sent Events.
// For each event, it parses the JSON object from the event data and appends the response to the fulltext string.
// If an error occurs during the POST request, the reading of the events, or the JSON unmarshalling, it fails the test.
func TestReadFromCloudflareLama2(t *testing.T) {
	// Retrieve the account ID and API token from the environment variables
	accountID := os.Getenv("CF_ACCOUNT_ID")
	apiToken := os.Getenv("CF_API_TOKEN")

	cf := &CloudflareAI{
		AccountID: accountID,
		APIToken:  apiToken,
	}

	// Send the POST request
	response, err := cf.Do("@cf/meta/llama-2-7b-chat-int8", &CfTextGenerationArg{
		Stream: true,
		Messages: []CfTextGenerationMsg{
			{Role: "system", Content: "You are a chatbot."},
			{Role: "user", Content: "What is your name?"},
		}})
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
		e := new(CfTextGenerationResponse)
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
