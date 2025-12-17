package gitcalendarcore

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/rdleal/intervalst/interval"
)

type (
	// The exposed API interface
	//
	// cannot expose channels, maps or some goofy types which do not have bindings to other languages
	Api interface {
		Initialize(repoPath string) error
		Clone(repoUrl, repoPath string) error
		// AddRemote()

		AddEvent(Event) error // TODO: check that it gets translated to a throwing exception for Kotlin/JS
		UpdateEvent(Event) error
		RemoveEvent(Event) error
		GetEvent(id int) (*Event, error)
		GetEvents(from int64, to int64) ([]*Event, error)
	}

	apiImpl struct {
		eventTree *interval.SearchTree[int, int64] // int: id; int64: timestamp end and start
		events    map[int]*Event
		repoPath  string
		repo      *git.Repository
	}
)

func NewApi() Api {
	var api apiImpl
	api.eventTree = interval.NewSearchTree[int](func(x, y int64) int { return int(x - y) })
	api.events = make(map[int]*Event)
	return &api
}

func (a *apiImpl) AddEvent(e Event) error {
	if err := e.Validate(); err != nil {
		return fmt.Errorf("invalid event data: %w", err)
	}

	// add to all events
	a.events[e.Id] = &e

	// -------- insert into tree --------
	err := a.eventTree.Insert(e.From, e.To, e.Id)
	if err != nil {
		return fmt.Errorf("failed to insert into index tree: %w", err)
	}

	// -------- create json file --------
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal event to JSON: %w", err)
	}

	filename := fmt.Sprintf("%d.json", e.Id)
	filePath := filepath.Join(a.repoPath, EventsDirName, filename)
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write event file: %w", err)
	}

	if a.repo == nil {
		return fmt.Errorf("repo not initialized")
	}

	// -------- add to git repo --------
	w, err := a.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	relativePath := filepath.Join(EventsDirName, filename)
	if _, err := w.Add(relativePath); err != nil {
		return fmt.Errorf("failed to stage event file: %w", err)
	}

	_, err = w.Commit(
		fmt.Sprintf("CALENDAR: Added event '%s'", e.Title),
		&git.CommitOptions{
			Author: &object.Signature{
				Name:  "git-calendar",
				Email: "",
				When:  time.Now(),
			},
		},
	)
	if err != nil {
		// TODO idk
		w.Remove(relativePath)
		return fmt.Errorf("failed to commit event: %w", err)
	}

	return err
}

func (a *apiImpl) UpdateEvent(e Event) error {
	if err := e.Validate(); err != nil {
		return fmt.Errorf("invalid event data: %w", err)
	}

	// check if it exists
	_, ok := a.events[e.Id]
	if !ok {
		return fmt.Errorf("event with this id doesnt exist")
	}

	// replace the pointer
	a.events[e.Id] = &e

	return nil
}

func (a *apiImpl) RemoveEvent(e Event) error {
	return nil
}

func (a *apiImpl) GetEvent(id int) (*Event, error) {
	e, ok := a.events[id]
	if !ok {
		return nil, fmt.Errorf("event with this id doesnt exist")
	}
	return e, nil
}

func (a *apiImpl) GetEvents(from int64, to int64) ([]*Event, error) {
	return nil, nil
}

func (a *apiImpl) Initialize(repoPath string) error {
	// ensure the base directory exists
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		return fmt.Errorf("failed to create repository directory: %w", err)
	}

	a.repoPath = repoPath

	// if it doesn't exist, initialize a new one
	repo, err := git.PlainInit(repoPath, false) // `false` for non-bare repo (has a worktree)

	if err == git.ErrTargetDirNotEmpty {
		repo, err := git.PlainOpen(repoPath)
		if err != nil {
			return fmt.Errorf("failed to open repository: %w", err)
		}
		a.repo = repo
		return nil

	} else if err != nil {
		return fmt.Errorf("failed to initialize new repository: %w", err)
	}

	a.repo = repo

	// create the events directory and an initial commit to ensure a master branch exists
	err = a.setupInitialRepoStructure(repoPath)
	return err
}

func (a *apiImpl) Clone(repoUrl, repoPath string) error {
	// check if the directory already exists and is non-empty
	if _, err := os.Stat(repoPath); err == nil {
		// if the directory exists, try to open it instead of cloning over it.
		// if the user meant to re-clone, they should delete the directory first.
		return a.Initialize(repoPath)
	}

	repo, err := git.PlainClone(repoPath, &git.CloneOptions{
		URL:      repoUrl,
		Progress: os.Stdout, // optional: for logging clone progress
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository from '%s': %w", repoUrl, err)
	}

	a.repo = repo
	return nil
}

func (a *apiImpl) setupInitialRepoStructure(repoPath string) error {
	err := os.Mkdir(path.Join(repoPath, EventsDirName), 0o755)
	if err != nil {
		return fmt.Errorf("failed to create folder '%s': %w", path.Join(repoPath, EventsDirName), err)
	}

	return nil
}
