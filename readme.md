[![GoDoc](https://pkg.go.dev/badge/github.com/mojocn/sseread?status.svg)](https://pkg.go.dev/github.com/mojocn/sseread?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/mojocn/sseread?)](https://goreportcard.com/report/github.com/mojocn/sseread)
[![codecov](https://codecov.io/gh/mojocn/sseread/branch/master/graph/badge.svg)](https://codecov.io/gh/mojocn/sseread)
[![Go version](https://img.shields.io/github/go-mod/go-version/mojocn/sseread.svg)](https://github.com/mojocn/sseread)
[![Follow mojocn](https://img.shields.io/github/followers/mojocn?label=Follow&style=social)](https://github.com/mojocn)


# Server Sent Events Reader

This is a simple library of how to read Server Sent Events (SSE) stream from `Response.Body` in Golang.


## Usage
download the library using
`go get -u github.com/mojocn/sseread@latest`

simple examples of how to use the library.

1. [example_test.go read by callback](/mojocn/sseread/blob/f8002c7d9655939755935a4ff143e01c8a67f583/example_test.go#L15) 
2. [example_test.go read by channel](/mojocn/sseread/blob/f8002c7d9655939755935a4ff143e01c8a67f583/example_test.go#L45)

```go
//sseread/example_test.go
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
```




## Testing

```bash
# git clone https://github.com/mojocn/sseread.git && cd sseread
go test -v
```





