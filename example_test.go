package sseread_test

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/mojocn/sseread"
)

func ExampleDo() {
	// Retrieve the account ID and API token from the environment variables
	accountID := os.Getenv("CF_ACCOUNT_ID")
	apiToken := os.Getenv("CF_API_TOKEN")

	cf := &sseread.CloudflareAI{
		AccountID: accountID,
		APIToken:  apiToken,
	}

	// Send the POST request
	response, err := cf.Do("@cf/meta/llama-2-7b-chat-fp8b", &sseread.CfTextGenerationArg{
		Stream: true,
		Messages: []sseread.CfTextGenerationMsg{
			{Role: "system", Content: "You are a chatbot."},
			{Role: "user", Content: "What is your name?"},
		}})
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
			return
		}
		log.Fatal(string(all))
		return
	}

	// Read the response body as Server-Sent Events
	channel, err := sseread.ReadCh(response.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Initialize an empty string to store the full text of the responses
	fulltext := ""

	// Iterate over the events from the channel
	for event := range channel {
		if event == nil || event.IsSkip() {
			continue
		}

		// Parse the JSON object from the event data
		e := new(sseread.CfTextGenerationResponse)
		err := json.Unmarshal(event.Data, e)
		if err != nil {
			log.Fatal(err, string(event.Data))
		} else {
			// Append the response to the fulltext string
			fulltext += e.Response
		}
	}

	// Log the full text of the responses
	fmt.Println(fulltext)
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
