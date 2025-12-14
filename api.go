package gitcalendarcore

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v6"
)

type (
	// The exposed API interface
	//
	// cannot expose channels or maps (they do not have a binding to other languages)
	Api interface {
		AddEvent(Event) error
		RemoveEvent(Event) error
		GetEvents(from time.Time, to time.Time) ([]Event, error) // TODO: check that it gets translated to a throwing exception for Kotlin/JS
	}

	apiImpl struct {
		repo *git.Repository
	}
)

func NewApi() Api {
	return &apiImpl{}
}

func (g *apiImpl) AddEvent(e Event) error {
	if err := e.Validate(); err != nil {
		return fmt.Errorf("invalid event data: %w", err)
	}
	return nil
}

func (g *apiImpl) RemoveEvent(e Event) error {
	return nil
}

func (g *apiImpl) GetEvents(from time.Time, to time.Time) ([]Event, error) {
	return nil, nil
}
