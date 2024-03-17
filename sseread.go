package sseread

import (
	"bufio"
	"bytes"
	"io"
	"log"
)

// Read reads from an io.Reader, parses the data as Server-Sent Events, and invokes the provided callback function for each event.
// It returns an error if any occurs during reading or parsing the events.
func Read(responseBody io.Reader, callback func(event *Event)) (err error) {
	scanner := bufio.NewScanner(responseBody)
	ev := new(Event)
	for scanner.Scan() {
		line := scanner.Bytes()
		firstColonIndex := bytes.IndexByte(line, ':')
		if firstColonIndex == -1 {
			callback(ev)
			//start another new server sent event
			ev = new(Event)
		} else {
			lineType, lineData := string(line[:firstColonIndex]), line[firstColonIndex+1:]
			//parse event filed(line)
			ev.ParseEventLine(lineType, lineData)
		}
	}
	return scanner.Err()
}

// ReadCh reads from an io.Reader, parses the data as Server-Sent Events, and sends each event on a channel.
// It returns the channel of events and an error if any occurs during reading or parsing the events.
func ReadCh(responseBody io.Reader) (messages <-chan *Event, err error) {
	channel := make(chan *Event)
	scanner := bufio.NewScanner(responseBody)
	go func() {
		defer close(channel)
		ev := new(Event)
		for scanner.Scan() {
			line := scanner.Bytes()
			log.Println(string(line))
			firstColonIndex := bytes.IndexByte(line, ':')
			if firstColonIndex == -1 {
				channel <- ev
				//start another new server sent event
				ev = new(Event)
			} else {
				lineType, lineData := string(line[:firstColonIndex]), line[firstColonIndex+1:]
				//parse event filed(line)
				ev.ParseEventLine(lineType, lineData)
			}
		}
	}()
	return channel, scanner.Err()
}
