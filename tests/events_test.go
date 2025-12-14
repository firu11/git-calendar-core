package tests

import (
	"encoding/json"
	"os"
	"path"
	"testing"
	"time"

	core "github.com/firu11/git-calendar-core"
)

func Test_AddEvent_CreatesJsonFile(t *testing.T) {
	a := core.NewApi()

	tmpDir := t.TempDir()
	err := a.Initialize(tmpDir)
	if err != nil {
		t.Errorf("failed to init repo: %v", err)
	}

	err = a.AddEvent(
		&core.Event{
			Id:    1,
			Title: "Foo Event",
			From:  time.Now().Unix(),
			To:    time.Now().Add(2 * time.Hour).Unix(),
		},
	)
	if err != nil {
		t.Errorf("failed to create an event: %v", err)
	}

	b, err := os.ReadFile(path.Join(tmpDir, core.EventsDirName, "1.json"))
	if err != nil {
		t.Errorf("failed to read event json file: %v", err)
	}

	var parsedEvent struct {
		Id    int    `json:"id"`
		Title string `json:"title"`
	}
	err = json.Unmarshal(b, &parsedEvent)
	if err != nil {
		t.Errorf("failed to parse event json file: %v", err)
	}

	if parsedEvent.Id != 1 {
		t.Errorf("id is not the same as input: \nin: %d\n !=\nfile: %v", 1, parsedEvent.Id)
	}
	if parsedEvent.Title != "Foo Event" {
		t.Errorf("id is not the same as input: \nin: %s\n !=\nfile: %s", "Foo Event", parsedEvent.Title)
	}
}
