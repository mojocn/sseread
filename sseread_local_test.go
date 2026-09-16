package sseread

import (
	"strings"
	"testing"
)

func TestReadParsesSSEFields(t *testing.T) {
	var events []*Event
	err := Read(strings.NewReader("id: 42\nevent: message\nretry: 1500\ndata: first\ndata: second\n\n: comment\nignored\n\n"), func(event *Event) {
		events = append(events, event)
	})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}
	if events[0].ID != "42" || events[0].Event != "message" || events[0].Retry != 1500 {
		t.Fatalf("unexpected event metadata: %+v", events[0])
	}
	if string(events[0].Data) != "first\nsecond" {
		t.Fatalf("data = %q, want first\\nsecond", events[0].Data)
	}
	if events[1].ID != "" || len(events[1].Data) != 0 {
		t.Fatalf("comment event should be empty: %+v", events[1])
	}
}

func TestReadChParsesEvents(t *testing.T) {
	messages, err := ReadCh(strings.NewReader("data: hello\n\ndata: world\n\n"))
	if err != nil {
		t.Fatalf("ReadCh returned error: %v", err)
	}

	var got []string
	for event := range messages {
		got = append(got, string(event.Data))
	}
	if len(got) != 2 || got[0] != "hello" || got[1] != "world" {
		t.Fatalf("events = %v, want [hello world]", got)
	}
}
