package sseread

import (
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
)

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
		t.Log(msg.ID, msg.Event, string(msg.Data))
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
		t.Log(msg.ID, msg.Event, string(msg.Data))
	}
	t.Logf("length of messages is %d", len(messages))

}

func safeClose(closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.Println("error closing:", err)
	}
}
