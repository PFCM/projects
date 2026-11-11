package main

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// TaskStatus is the state that a task can be in.
type TaskStatus byte

const (
	// TaskStatusUnknown is an invalid or unknown TaskStatus.
	TaskStatusUnknown TaskStatus = iota
	// TaskStatusTODO is the status of a task that has not been completed,
	// and no one is working on. If it's also not blocked, that means it's
	// ready to be picked up.
	TaskStatusTODO
	// TaskStatusInProgress is the status of a task that someone is working
	// on. Not available to be picked up.
	TaskStatusInProgress
	// TaskStatusInReview is the status of a taks awaiting review. It must
	// have a review file, or something has gone wrong.
	TaskStatusInReview
	// TaskStatusDone is a task that is finished.
	TaskStatusDone
)

var taskStatusToString = []string{
	TaskStatusUnknown:    "Unknown",
	TaskStatusTODO:       "todo",
	TaskStatusInProgress: "in-progress",
	TaskStatusInReview:   "in-review",
	TaskStatusDone:       "done",
}

var taskStatusFromString = func() map[string]TaskStatus {
	m := make(map[string]TaskStatus, len(taskStatusToString))
	for i, s := range taskStatusToString {
		m[s] = TaskStatus(i)
	}
	return m
}()

func (t TaskStatus) String() string {
	if t < 0 || int(t) >= len(taskStatusToString) {
		return "INVALID"
	}
	return taskStatusToString[t]
}

func (t TaskStatus) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, t.String())
}

func (t *TaskStatus) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var raw string
	if err := json.UnmarshalDecode(dec, &raw); err != nil {
		return err
	}
	ts, ok := taskStatusFromString[raw]
	if !ok {
		return fmt.Errorf("unknown TaskStatus: %v", raw)
	}
	*t = ts
	return nil
}

// Task is a single piece of work.
type Task struct {
	// ID uniquely identifies the task within the project.
	ID string `json:"id"`
	// Title is a short, human readable name for the task.
	Title string `json:"title"`
	// Goal is a longer description of the task framed around the outcome:
	// what will be true when this task is complete.
	Goal string `json:"goal"`
	// References provide pointers to context necessary to complete the
	// task: documentation links, existing code, etc.
	References []string `json:"references"`
	// Criteria contains the functional requirements that must be true to
	// consider the goal satisfied, and describe how the task will be judged
	// to be complete.
	Criteria []string `json:"criteria"`
	// Expectations describe what a good solution looks like. As opposed to
	// criteria these are non-functional requirements, style guidelines and
	// so on.
	Expectations []string `json:"expectations"`
	// BlockedBy lists the ids of tasks that must be complete before this
	// task can be started.
	BlockedBy []string `json:"blocked_by,omitempty"`
	// Status is the status of the task, see TaskStatus for values.
	Status TaskStatus `json:"status"`
	// Claim, if non-zero, is the details of who has claimed to be working
	// on the task.
	Claim Claim `json:"claim,omitzero"`
}

// Claim is who took a task and when.
type Claim struct {
	// By is a username or claude:<session id> for a claude agent.
	By string `json:"by"`
	// At is when the task was claimed.
	At time.Time `json:"at"`
}

func readTask(path string) (Task, error) {
	f, err := os.Open(path)
	if err != nil {
		return Task{}, err
	}
	defer f.Close()
	var (
		t   Task
		dec = jsontext.NewDecoder(f)
	)
	if err := json.UnmarshalDecode(dec, &t); err != nil {
		return Task{}, err
	}
	if _, err := dec.ReadToken(); !errors.Is(err, io.EOF) {
		if err == nil {
			return Task{}, errors.New("trailing content")
		}
		return Task{}, err
	}
	// TODO: more thorough validation
	if t.Status == TaskStatusUnknown {
		return Task{}, errors.New("no status provided")
	}

	return t, nil
}

func readTasks(dir string) ([]Task, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "tasks", "*", "task.json"))
	if err != nil {
		return nil, err
	}
	var tasks []Task
	for _, path := range paths {
		t, err := readTask(path)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}
