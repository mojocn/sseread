package sseread

import (
	"bufio"
	"bytes"
	"io"
)

func scanEvents(responseBody io.Reader, callback func(event *Event)) error {
	scanner := bufio.NewScanner(responseBody)
	ev := new(Event)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			callback(ev)
			ev = new(Event)
			continue
		}
		firstColonIndex := bytes.IndexByte(line, ':')
		if firstColonIndex == -1 {
			continue
		}
		ev.ParseEventLine(string(line[:firstColonIndex]), line[firstColonIndex+1:])
	}
	return scanner.Err()
}

// Read reads from an io.Reader, parses the data as Server-Sent Events, and invokes the provided callback function for each event.
// It returns an error if any occurs during reading or parsing the events.
func Read(responseBody io.Reader, callback func(event *Event)) (err error) {
	return scanEvents(responseBody, callback)
}

// ReadCh reads from an io.Reader, parses the data as Server-Sent Events, and sends each event on a channel.
// It returns the channel of events and an error if any occurs during reading or parsing the events.
func ReadCh(responseBody io.Reader) (messages <-chan *Event, err error) {
	channel := make(chan *Event)
	go func() {
		defer close(channel)
		_ = scanEvents(responseBody, func(event *Event) { channel <- event })
	}()
	return channel, nil
}
