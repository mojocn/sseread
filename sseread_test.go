package sseread

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
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

// TestRead is a test function for the Read function in the sseread package.
// It sends a GET request to the specified URL and reads the response body as Server-Sent Events.
// For each event, it appends it to the messages slice and logs the event ID, event type, and event data.
// If an error occurs during the GET request or the reading of the events, it fails the test.
func TestRead(t *testing.T) {
	response, err := http.Get("https://mojotv.cn/api/sse") //API source code from https://github.com/mojocn/gptchat/blob/main/app/api/sse/route.ts
	if err != nil {
		t.Fatal(err)
	}

	ct := strings.ToLower(response.Header.Get("Content-Type"))
	if !strings.Contains(ct, "text/event-stream") {
		t.Fatal("expect content-type: text/event-stream, but actual", ct)
	}

	defer safeClose(response.Body) //don't forget to close the response body

	var messages []Event
	err = Read(response.Body, func(msg *Event) {
		if msg != nil {
			messages = append(messages, *msg)
		}
		t.Log(msg.ID, msg.Event, string(msg.Data), msg.Retry)
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("length of messages is %d", len(messages))
}

// TestReadChannel is a test function for the ReadCh function in the sseread package.
// It sends a GET request to the specified URL and reads the response body as Server-Sent Events.
// For each event, it logs the event ID, event type, and event data.
// If an error occurs during the GET request or the reading of the events, it fails the test.
func TestReadChannel(t *testing.T) {
	response, err := http.Get("https://mojotv.cn/api/sse") //API source code from https://github.com/mojocn/gptchat/blob/main/app/api/sse/route.ts
	if err != nil {
		t.Fatal(err)
	}
	defer safeClose(response.Body) //don't forget to close the response body

	ct := strings.ToLower(response.Header.Get("Content-Type"))
	if !strings.Contains(ct, "text/event-stream") {
		t.Fatal("expect content-type: text/event-stream, but actual", ct)
	}

	channel, err := ReadCh(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	var messages []Event
	for msg := range channel {
		if msg != nil {
			messages = append(messages, *msg)
		}
		t.Log(msg.ID, msg.Event, string(msg.Data), msg.Retry)
	}
	t.Logf("length of messages is %d", len(messages))

}

func safeClose(closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.Println("error closing:", err)
	}
}
