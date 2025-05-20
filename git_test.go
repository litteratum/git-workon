package main

import (
	"errors"
	"testing"
)

func TestClone(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		cmd := &FakeCMD{}
		git := NewGitAPI(cmd)

		source := "source"
		destination := "dest"
		err := git.Clone(source, destination)
		assertNotError(t, err)
		assertHistory(
			t,
			cmd,
			[]map[string]any{
				{
					"_method": "Run",
					"name":    "git",
					"args":    []string{"clone", source, destination},
				},
			},
		)
	})
	t.Run("cmd error", func(t *testing.T) {
		cmd := &FakeCMD{
			err: errors.New("cmd err"),
		}
		git := NewGitAPI(cmd)
		err := git.Clone("s", "d")
		assertError(t, err)
	})
}

func TestGetProjectState(t *testing.T) {
	t.Run("ok; clean", func(t *testing.T) {
		cmd := &FakeCMD{}
		git := NewGitAPI(cmd)

		state, err := git.GetProjectState("proj/path")
		assertNotError(t, err)
		assertHistory(
			t,
			cmd,
			[]map[string]any{
				{
					"_method": "RunCwd",
					"dir":     "proj/path",
					"name":    "git",
					"args":    []string{"stash", "list"},
				},
				{
					"_method": "RunCwd",
					"dir":     "proj/path",
					"name":    "git",
					"args":    []string{"push", "--tags", "--dry-run"},
				},
				{
					"_method": "RunCwd",
					"dir":     "proj/path",
					"name":    "git",
					"args":    []string{"log", "--branches", "--not", "--remotes", "--decorate", "--oneline"},
				},
				{
					"_method": "RunCwd",
					"dir":     "proj/path",
					"name":    "git",
					"args":    []string{"status", "--short"},
				},
			},
		)

		if !state.Clean() {
			t.Fatalf("expected clean state, got %v", state)
		}
	})

	tests := map[string]struct {
		cmdResults []CMDResult
	}{
		"stashes": {
			cmdResults: []CMDResult{
				{Stdout: "stash"},
			},
		},
		"tags": {
			cmdResults: []CMDResult{
				{},
				{Stderr: "new tag"},
			},
		},
		"commits": {
			cmdResults: []CMDResult{
				{},
				{},
				{Stdout: "commit"},
			},
		},
		"status": {
			cmdResults: []CMDResult{
				{},
				{},
				{},
				{Stdout: "status"},
			},
		},
		"mixed": {
			cmdResults: []CMDResult{
				{Stdout: "stash"},
				{Stderr: "new tag"},
				{Stdout: "commit"},
				{Stdout: "status"},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := &FakeCMD{
				results: test.cmdResults,
			}
			git := NewGitAPI(cmd)

			state, err := git.GetProjectState("proj/path")
			assertNotError(t, err)
			if state.Clean() {
				t.Fatalf("expected not clean state, got %v", state)
			}
		})
	}
}
