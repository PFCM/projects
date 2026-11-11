package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestReadTask(t *testing.T) {
	for _, c := range []struct {
		name    string
		in      string
		want    Task
		wantErr bool
	}{{
		name:    "empty",
		in:      "",
		wantErr: true,
	}, {
		name:    "malformed",
		in:      `{"id": "T1",`,
		wantErr: true,
	}, {
		name:    "trailing-data",
		in:      `{"id": "T1", "status": "todo"} {"id": "T2", "status": "todo"}`,
		wantErr: true,
	}, {
		name:    "duplicate-key",
		in:      `{"id": "T1", "id": "T2", "status": "todo"}`,
		wantErr: true,
	}, {
		name: "unknown-field",
		in:   `{"id": "T1", "status": "todo", "colour": "red"}`,
		want: Task{ID: "T1", Status: TaskStatusTODO},
	}, {
		name:    "status-missing",
		in:      `{"id": "T1"}`,
		wantErr: true,
	}, {
		name:    "status-null",
		in:      `{"id": "T1", "status": null}`,
		wantErr: true,
	}, {
		name:    "status-unknown",
		in:      `{"id": "T1", "status": "doing"}`,
		wantErr: true,
	}, {
		name: "claim-null",
		in:   `{"id": "T1", "status": "todo", "claim": null}`,
		want: Task{ID: "T1", Status: TaskStatusTODO},
	}, {
		name: "claim-empty",
		in:   `{"id": "T1", "status": "todo", "claim": {}}`,
		want: Task{ID: "T1", Status: TaskStatusTODO},
	}, {
		name:    "claim-wrong-type",
		in:      `{"id": "T1", "status": "todo", "claim": "paul"}`,
		wantErr: true,
	}, {
		name:    "claim-bad-time",
		in:      `{"id": "T1", "status": "in-progress", "claim": {"by": "paul", "at": "yesterday"}}`,
		wantErr: true,
	}, {
		name: "unclaimed",
		in:   `{"id": "T1", "title": "Create the repo", "blocked_by": ["T2", "T10"], "status": "todo"}`,
		want: Task{ID: "T1", Title: "Create the repo", BlockedBy: []string{"T2", "T10"}, Status: TaskStatusTODO},
	}, {
		name: "claimed",
		in:   `{"id": "T1", "status": "in-progress", "claim": {"by": "paul", "at": "2026-11-11T10:02:00Z"}}`,
		want: Task{ID: "T1", Status: TaskStatusInProgress, Claim: Claim{By: "paul", At: time.Date(2026, 11, 11, 10, 2, 0, 0, time.UTC)}},
	}} {
		t.Run(c.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "task.json")
			if err := os.WriteFile(path, []byte(c.in), 0o644); err != nil {
				t.Fatal(err)
			}
			got, err := readTask(path)
			if c.wantErr {
				if err == nil {
					t.Errorf("readTask(%s) = %+v, want an error", c.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("readTask(%s): unexpected error: %v", c.in, err)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("readTask(%s):\n got  %+v\n want %+v", c.in, got, c.want)
			}
		})
	}
}

func TestReadTaskMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "task.json")
	if got, err := readTask(path); err == nil {
		t.Errorf("readTask of a missing file = %+v, want an error", got)
	}
}
