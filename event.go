package gitcalendarcore

import "errors"

type Event struct {
	Id              int
	Title           string
	Location        string
	FromTimestampMS int64 // not using time.Time for cross-lang compatibility
	ToTimestampMS   int64
}

func (e *Event) Validate() error {
	if e.Title == "" {
		return errors.New("event title cannot be empty")
	}
	if e.FromTimestampMS < 0 || e.ToTimestampMS < 0 {
		return errors.New("event timestamps cannot be before epoch (less than 0)")
	}
	if e.FromTimestampMS >= e.ToTimestampMS {
		return errors.New("event `from` timestamp cannot be greater or equal than `to` (cannot end before it starts)")
	}
	return nil
}
