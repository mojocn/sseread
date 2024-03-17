package sseread_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mojocn/sseread"
	"io"
	"net/http"
	"strings"
)

// ExampleReadCloudflareAI is a function that demonstrates how to interact with the Cloudflare AI API.
// It sends a POST request to the API and reads the response body as Server-Sent Events (SSE).
// For each event, it parses the JSON object from the event data and appends the response to the fulltext string.
// If an error occurs during the POST request, the reading of the events, or the JSON unmarshalling, it prints the error and returns from the function.
func ExampleReadCloudflareAI() {
	// Replace these with your actual account ID and API token
	accountID := "xxxx"
	apiToken := "yyy"

	// Create a buffer with the request body
	buff := bytes.NewBufferString(`{ "stream":true,"messages": [{ "role": "system", "content": "You are a friendly assistant" }, { "role": "user", "content": "Why is pizza so good" }]}`)

	// Create a new POST request to the Cloudflare AI API
	req, err := http.NewRequest("POST", "https://api.cloudflare.com/client/v4/accounts/"+accountID+"/ai/run/@cf/meta/llama-2-7b-chat-int8", buff)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Set the Authorization header with the API token
	req.Header.Set("Authorization", "Bearer "+apiToken)

	// Send the POST request
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Ensure the response body is closed after the function returns
	defer response.Body.Close()

	// Check the response status code
	if response.StatusCode != http.StatusOK {
		all, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(all))
		fmt.Println("response status code is not 200")
		return
	}

	// Read the response body as Server-Sent Events
	channel, err := sseread.ReadCh(response.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Define a struct to unmarshal the JSON object from the event data
	type llamaMsg struct {
		Response string `json:"response"`
		P        string `json:"p"`
	}

	// Initialize an empty string to store the full text of the responses
	fulltext := ""

	// Iterate over the events from the channel
	for event := range channel {
		// If the event is nil, skip it
		if event == nil {
			continue
		}

		// Parse the JSON object from the event data
		e := new(llamaMsg)
		err := json.Unmarshal(event.Data, e)
		if err != nil {
			fmt.Println(err)
		} else {
			// Append the response to the fulltext string
			fulltext += e.Response
		}
	}

	// Print the full text of the responses
	fmt.Printf(fulltext)
}

// ExampleRead is a function that demonstrates how to read Server-Sent Events (SSE) from a specific URL.
func ExampleRead() {
	// Send a GET request to the specified URL.
	response, err := http.Get("https://mojotv.cn/api/sse") //API source code from https://github.com/mojocn/gptchat/blob/main/app/api/sse/route.ts
	// If an error occurs during the GET request, print the error and return from the function.
	if err != nil {
		fmt.Println(err)
		return
	}

	// Get the Content-Type header from the response and convert it to lowercase.
	ct := strings.ToLower(response.Header.Get("Content-Type"))
	// If the Content-Type is not "text/event-stream", print a message and continue.
	if !strings.Contains(ct, "text/event-stream") {
		fmt.Println("expect content-type: text/event-stream, but actual", ct)
	}

	// Ensure the response body is closed when the function returns.
	defer response.Body.Close() //don't forget to close the response body

	// Declare a slice to store the SSE messages.
	var messages []sseread.Event
	// Read the SSE messages from the response body.
	err = sseread.Read(response.Body, func(msg *sseread.Event) {
		// If the message is not nil, append it to the messages slice.
		if msg != nil {
			messages = append(messages, *msg)
		}
		// Print the ID, event type, and data of the message.
		fmt.Println(msg.ID, msg.Event, string(msg.Data))
	})
	// If an error occurs while reading the SSE messages, print the error and return from the function.
	if err != nil {
		fmt.Println(err)
		return
	}
	// Print the number of messages read.
	fmt.Printf("length of messages is %d", len(messages))
}

// ExampleReadCh is a function that demonstrates how to read Server-Sent Events (SSE) from a specific URL using channels.
func ExampleReadCh() {
	// Send a GET request to the specified URL.
	response, err := http.Get("https://mojotv.cn/api/sse") //API source code from https://github.com/mojocn/gptchat/blob/main/app/api/sse/route.ts
	// If an error occurs during the GET request, print the error and return from the function.
	if err != nil {
		fmt.Println(err)
		return
	}

	// Get the Content-Type header from the response and convert it to lowercase.
	ct := strings.ToLower(response.Header.Get("Content-Type"))
	// If the Content-Type is not "text/event-stream", print a message and continue.
	if !strings.Contains(ct, "text/event-stream") {
		fmt.Println("expect content-type: text/event-stream, but actual", ct)
	}

	// Ensure the response body is closed when the function returns.
	defer response.Body.Close() //don't forget to close the response body

	// Read the SSE messages from the response body into a channel.
	channel, err := sseread.ReadCh(response.Body)
	// If an error occurs while reading the SSE messages, print the error and return from the function.
	if err != nil {
		fmt.Println(err)
		return
	}
	// Declare a slice to store the SSE messages.
	var messages []sseread.Event
	// Loop over the channel to receive the SSE messages.
	for msg := range channel {
		// If the message is not nil, append it to the messages slice.
		if msg != nil {
			messages = append(messages, *msg)
		}
		// Print the ID, event type, and data of the message.
		fmt.Println(msg.ID, msg.Event, string(msg.Data))
	}
	// Print the number of messages read.
	fmt.Printf("length of messages is %d", len(messages))
}
