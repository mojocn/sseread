package sseread

import (
	"encoding/json"
	"strings"
)

// EventLineParser is an interface that defines a method for parsing event fields.
type EventLineParser interface {
	// ParseEventLine is a method that takes an event type and event data as input and parses the event field.
	ParseEventLine(lineType string, lineData []byte)
}

// Event represents a Server-Sent Event.
// It contains fields for the event ID, retry count, event type, and event data.
type Event struct {
	ID    string          // ID is the unique identifier for the event.
	Retry uint            // Retry is the number of times the event should be retried.
	Event string          // Event is the type of the event.
	Data  json.RawMessage // Data is the raw JSON data associated with the event. Or just convert it to string.
}

// ParseEventLine is a method of the Event struct that parses an event field based on the event type.
// It takes an event type and event data as input, and updates the corresponding field in the Event struct.
func (e *Event) ParseEventLine(lineType string, lineData []byte) {

	switch strings.TrimSpace(lineType) {
	case "event":
		e.Event = string(lineData) // If the event type is "event", update the Event field.
	case "id":
		e.ID = string(lineData) // If the event type is "id", update the ID field.
	case "retry":
		e.ID = string(lineData) // If the event type is "retry", update the Retry field.
	case "data":
		e.Data = lineData // If the event type is "data", update the Data field.

	}

}

// IsSkip is a method of the Event struct that checks if the event should be skipped.
// It returns true if the Data field of the event is empty, null, undefined, or "[DONE]".
// Otherwise, it returns false.
func (e *Event) IsSkip() bool {
	// Check if the Data field is empty.
	if len(e.Data) == 0 {
		return true
	}
	// Convert the Data field to a string and remove leading and trailing spaces.
	str := strings.TrimSpace(string(e.Data))
	// Check if the Data field is "", "null", "undefined", or "[DONE]".
	if str == "" || str == "null" || str == "undefined" || str == "[DONE]" {
		return true
	}
	// If none of the above conditions are met, return false.
	return false
}
