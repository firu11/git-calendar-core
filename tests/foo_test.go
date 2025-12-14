package main

import (
	"testing"
	"time"

	core "github.com/firu11/git-calendar-core"
)

func TestFoo(t *testing.T) {
	g := core.NewApi()

	err := g.AddEvent(
		core.Event{
			Title:           "FooEvent",
			FromTimestampMS: time.Now().Unix(),
			ToTimestampMS:   time.Now().Add(2 * time.Hour).Unix(),
		},
	)
	if err != nil {
		t.Errorf("failed to create an event: %v", err)
	}
}
